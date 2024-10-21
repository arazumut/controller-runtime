/*
2021 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
haklar ve sınırlamalar için Lisans'a bakınız.
*/

package certwatcher

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/certwatcher/metrics"
	logf "sigs.k8s.io/controller-runtime/pkg/internal/log"
)

var log = logf.RuntimeLog.WithName("certwatcher")

// CertWatcher, sertifika ve anahtar dosyalarını değişiklikler için izler.
// Her iki dosya değiştiğinde, her ikisini de okur ve ayrıştırır ve yeni sertifika ile isteğe bağlı bir geri çağırma işlevi çağırır.
type CertWatcher struct {
	sync.RWMutex

	currentCert *tls.Certificate
	watcher     *fsnotify.Watcher

	certPath string
	keyPath  string

	// callback, sertifika değiştiğinde çağrılacak bir işlevdir.
	callback func(tls.Certificate)
}

// Yeni bir CertWatcher döndürür ve belirtilen sertifika ve anahtarı izler.
func New(certPath, keyPath string) (*CertWatcher, error) {
	var err error

	cw := &CertWatcher{
		certPath: certPath,
		keyPath:  keyPath,
	}

	// Sertifika ve anahtarın ilk okunması.
	if err := cw.ReadCertificate(); err != nil {
		return nil, err
	}

	cw.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return cw, nil
}

// RegisterCallback, sertifika değiştiğinde çağrılacak bir geri çağırma işlevi kaydeder.
func (cw *CertWatcher) RegisterCallback(callback func(tls.Certificate)) {
	cw.Lock()
	defer cw.Unlock()
	// Mevcut sertifika null değilse, geri çağırmayı hemen çağır.
	if cw.currentCert != nil {
		callback(*cw.currentCert)
	}
	cw.callback = callback
}

// GetCertificate, şu anda yüklenmiş olan sertifikayı alır, bu null olabilir.
func (cw *CertWatcher) GetCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	cw.RLock()
	defer cw.RUnlock()
	return cw.currentCert, nil
}

// Sertifika ve anahtar dosyaları üzerinde izlemeyi başlatır.
func (cw *CertWatcher) Start(ctx context.Context) error {
	files := sets.New(cw.certPath, cw.keyPath)

	{
		var watchErr error
		if err := wait.PollUntilContextTimeout(ctx, 1*time.Second, 10*time.Second, true, func(ctx context.Context) (done bool, err error) {
			for _, f := range files.UnsortedList() {
				if err := cw.watcher.Add(f); err != nil {
					watchErr = err
					return false, nil //nolint:nilerr // Denemeye devam etmek istiyoruz.
				}
				// İzlemeyi ekledik, setten çıkar.
				files.Delete(f)
			}
			return true, nil
		}); err != nil {
			return fmt.Errorf("izlemeler eklenemedi: %w", kerrors.NewAggregate([]error{err, watchErr}))
		}
	}

	go cw.Watch()

	log.Info("Sertifika izleyici başlatılıyor")

	// Bağlam tamamlanana kadar bekle.
	<-ctx.Done()

	return cw.watcher.Close()
}

// İzleyicinin kanalından olayları okur ve değişikliklere tepki verir.
func (cw *CertWatcher) Watch() {
	for {
		select {
		case event, ok := <-cw.watcher.Events:
			// Kanal kapalı.
			if !ok {
				return
			}

			cw.handleEvent(event)

		case err, ok := <-cw.watcher.Errors:
			// Kanal kapalı.
			if !ok {
				return
			}

			log.Error(err, "sertifika izleme hatası")
		}
	}
}

// Sertifika ve anahtar dosyalarını diskten okur, ayrıştırır ve izleyicideki mevcut sertifikayı günceller.
// Bir geri çağırma ayarlanmışsa, yeni sertifika ile çağrılır.
func (cw *CertWatcher) ReadCertificate() error {
	metrics.ReadCertificateTotal.Inc()
	cert, err := tls.LoadX509KeyPair(cw.certPath, cw.keyPath)
	if err != nil {
		metrics.ReadCertificateErrors.Inc()
		return err
	}

	cw.Lock()
	cw.currentCert = &cert
	cw.Unlock()

	log.Info("Mevcut TLS sertifikası güncellendi")

	// Bir geri çağırma kaydedilmişse, yeni sertifika ile çağır.
	cw.RLock()
	defer cw.RUnlock()
	if cw.callback != nil {
		go func() {
			cw.callback(cert)
		}()
	}
	return nil
}

func (cw *CertWatcher) handleEvent(event fsnotify.Event) {
	// Yalnızca dosyanın içeriğini değiştirebilecek olaylarla ilgilenir.
	if !(isWrite(event) || isRemove(event) || isCreate(event) || isChmod(event)) {
		return
	}

	log.V(1).Info("sertifika olayı", "event", event)

	// Dosya kaldırıldı veya yeniden adlandırıldıysa, önceki ada izlemeyi yeniden ekle
	if isRemove(event) || isChmod(event) {
		if err := cw.watcher.Add(event.Name); err != nil {
			log.Error(err, "dosya yeniden izlenirken hata")
		}
	}

	if err := cw.ReadCertificate(); err != nil {
		log.Error(err, "sertifika yeniden okunurken hata")
	}
}

func isWrite(event fsnotify.Event) bool {
	return event.Op.Has(fsnotify.Write)
}

func isCreate(event fsnotify.Event) bool {
	return event.Op.Has(fsnotify.Create)
}

func isRemove(event fsnotify.Event) bool {
	return event.Op.Has(fsnotify.Remove)
}

func isChmod(event fsnotify.Event) bool {
	return event.Op.Has(fsnotify.Chmod)
}
