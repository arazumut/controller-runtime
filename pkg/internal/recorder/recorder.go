/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") altında lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
herhangi bir garanti veya koşul olmaksızın, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisansa bakınız.
*/

package recorder

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
)

// EventBroadcasterProducer bir olay yayıncısı oluşturur ve
// yayıncının Sağlayıcı ile durdurulup durdurulmayacağını döndürür,
// veya değil (örneğin, paylaşılıyorsa, Sağlayıcı ile durdurulmamalıdır).
type EventBroadcasterProducer func() (caster record.EventBroadcaster, stopWithProvider bool)

// Provider, olayları k8s API sunucusuna ve bir logr Logger'a kaydeden bir recorder.Provider'dır.
type Provider struct {
	lock    sync.RWMutex
	stopped bool

	// kaydedici oluştururken belirtilecek şema
	scheme *runtime.Scheme
	// tanılama olay bilgilerini kaydederken kullanılacak logger
	logger          logr.Logger
	evtClient       corev1client.EventInterface
	makeBroadcaster EventBroadcasterProducer

	broadcasterOnce sync.Once
	broadcaster     record.EventBroadcaster
	stopBroadcaster bool
}

// NB(directxman12): bu, bir runnable olmaktan ziyade Stop'u manuel olarak uygular çünkü
// her şey kapandıktan *sonra* durdurulması gerekir, aksi takdirde liderlik seçimi
// kodu bitip olayları yaymaya devam etmeye çalışırken paniklere neden olur.

// Stop, bu sağlayıcıyı durdurmaya çalışır, alttaki yayıncıyı durdurulması istenirse durdurur.
// Verilen bağlamı onurlandırmaya çalışır, ancak alttaki yayıncı, tüm sıradaki olaylar
// temizlenene kadar dönmeyen belirsiz bir bekleme süresine sahiptir, bu nedenle bu, alttaki
// bekleme süresi bitmeden önce dönmek yerine beklemeyi iptal edebilir.
// Bu Çok Sinir Bozucu™.
func (p *Provider) Stop(shutdownCtx context.Context) {
	doneCh := make(chan struct{})

	go func() {
		// teknik olarak, bu yayıncıyı başlatabilir, ancak pratikte, neredeyse kesinlikle
		// zaten başlatılmıştır (örneğin, liderlik seçimi tarafından). Bunu çağırmamız
		// gerekiyor ki getBroadcaster çağrısıyla yarışmayalım.
		broadcaster := p.getBroadcaster()
		if p.stopBroadcaster {
			p.lock.Lock()
			broadcaster.Shutdown()
			p.stopped = true
			p.lock.Unlock()
		}
		close(doneCh)
	}()

	select {
	case <-shutdownCtx.Done():
	case <-doneCh:
	}
}

// getBroadcaster, bu sağlayıcı için bir yayıncının başlatıldığından emin olur ve onu döndürür.
// Bu iş parçacığı güvenlidir.
func (p *Provider) getBroadcaster() record.EventBroadcaster {
	// NB(directxman12): bu, birisi "getBroadcaster" çağırırsa (yani bir Olay Yayar)
	// ancak Start'ı çağırmazsa teknik olarak sızabilir, ancak yayıncıyı başlatmada
	// yarışabiliriz. Alternatif, olayları sessizce yutmak ve daha fazla kilitleme,
	// ancak bu altoptimal görünüyor.

	p.broadcasterOnce.Do(func() {
		broadcaster, stop := p.makeBroadcaster()
		broadcaster.StartRecordingToSink(&corev1client.EventSinkImpl{Interface: p.evtClient})
		broadcaster.StartEventWatcher(
			func(e *corev1.Event) {
				p.logger.V(1).Info(e.Message, "type", e.Type, "object", e.InvolvedObject, "reason", e.Reason)
			})
		p.broadcaster = broadcaster
		p.stopBroadcaster = stop
	})

	return p.broadcaster
}

// NewProvider yeni bir Provider örneği oluşturur.
func NewProvider(config *rest.Config, httpClient *http.Client, scheme *runtime.Scheme, logger logr.Logger, makeBroadcaster EventBroadcasterProducer) (*Provider, error) {
	if httpClient == nil {
		panic("httpClient boş olmamalıdır")
	}

	corev1Client, err := corev1client.NewForConfigAndClient(config, httpClient)
	if err != nil {
		return nil, fmt.Errorf("istemci başlatılamadı: %w", err)
	}

	p := &Provider{scheme: scheme, logger: logger, makeBroadcaster: makeBroadcaster, evtClient: corev1Client.Events("")}
	return p, nil
}

// GetEventRecorderFor, bu sağlayıcının yayıncısına yayın yapan bir olay kaydedici döndürür.
// Tüm olaylar verilen adın bir bileşeni ile ilişkilendirilecektir.
func (p *Provider) GetEventRecorderFor(name string) record.EventRecorder {
	return &lazyRecorder{
		prov: p,
		name: name,
	}
}

// lazyRecorder, alttaki kaydedici aslında ilk olay yayılana kadar herhangi bir kaydedici
// oluşturmaz.
type lazyRecorder struct {
	prov *Provider
	name string

	recOnce sync.Once
	rec     record.EventRecorder
}

// ensureRecording, bu kaydedici için somut bir kaydedicinin doldurulmasını sağlar.
func (l *lazyRecorder) ensureRecording() {
	l.recOnce.Do(func() {
		broadcaster := l.prov.getBroadcaster()
		l.rec = broadcaster.NewRecorder(l.prov.scheme, corev1.EventSource{Component: l.name})
	})
}

func (l *lazyRecorder) Event(object runtime.Object, eventtype, reason, message string) {
	l.ensureRecording()

	l.prov.lock.RLock()
	if !l.prov.stopped {
		l.rec.Event(object, eventtype, reason, message)
	}
	l.prov.lock.RUnlock()
}
func (l *lazyRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
	l.ensureRecording()

	l.prov.lock.RLock()
	if !l.prov.stopped {
		l.rec.Eventf(object, eventtype, reason, messageFmt, args...)
	}
	l.prov.lock.RUnlock()
}
func (l *lazyRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
	l.ensureRecording()

	l.prov.lock.RLock()
	if !l.prov.stopped {
		l.rec.AnnotatedEventf(object, annotations, eventtype, reason, messageFmt, args...)
	}
	l.prov.lock.RUnlock()
}
