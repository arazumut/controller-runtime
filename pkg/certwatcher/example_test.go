/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
yetkiler ve sınırlamalar için Lisansa bakınız.
*/

package certwatcher_test

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
)

type ornekSunucu struct {
}

func Ornek() {
	// Context'i ayarla
	ctx := ctrl.SetupSignalHandler()

	// Yeni bir sertifika izleyici başlat
	watcher, err := certwatcher.New("ssl/tls.crt", "ssl/tls.key")
	if err != nil {
		panic(err)
	}

	// Sertifika izleyiciyi çalıştıran bir goroutine başlat
	go func() {
		if err := watcher.Start(ctx); err != nil {
			panic(err)
		}
	}()

	// Sertifika değişikliklerinde sertifikayı almak için GetCertificate kullanarak TLS dinleyiciyi ayarla
	listener, err := tls.Listen("tcp", "localhost:9443", &tls.Config{
		GetCertificate: watcher.GetCertificate,
		MinVersion:     tls.VersionTLS12,
	})
	if err != nil {
		panic(err)
	}

	// TLS sunucusunu başlat
	srv := &http.Server{
		Handler:           &ornekSunucu{},
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Sunucu kapatmayı yöneten bir goroutine başlat
	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			panic(err)
		}
	}()

	// Sunucuyu başlat
	if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

func (s *ornekSunucu) ServeHTTP(http.ResponseWriter, *http.Request) {
}
