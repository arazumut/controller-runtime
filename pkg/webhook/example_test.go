/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa veya yazılı izin gerektirmedikçe, bu yazılım
Lisans kapsamında "OLDUĞU GİBİ" dağıtılmaktadır,
herhangi bir garanti veya koşul olmaksızın.
Lisans kapsamında izin verilen belirli dil kapsamındaki
haklar ve sınırlamalar için Lisansa bakınız.
*/

package webhook_test

import (
	"context"
	"net/http"

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/internal/log"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	. "sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	// Çeşitli sunucu yapılandırma seçenekleri için kullanılan web kancalarını oluşturun
	//
	// Bu işleyiciler, daha karmaşık uygulamalar için
	// AdmissionHandler arayüzünün uygulamaları da olabilir.
	mutatingHook = &Admission{
		Handler: admission.HandlerFunc(func(ctx context.Context, req AdmissionRequest) AdmissionResponse {
			return Patched("bazı değişiklikler",
				JSONPatchOp{Operation: "add", Path: "/metadata/annotations/access", Value: "granted"},
				JSONPatchOp{Operation: "add", Path: "/metadata/annotations/reason", Value: "not so secret"},
			)
		}),
	}

	validatingHook = &Admission{
		Handler: admission.HandlerFunc(func(ctx context.Context, req AdmissionRequest) AdmissionResponse {
			return Denied("hiç kimse geçemez!")
		}),
	}
)

// Bu örnek, bir denetleyici yöneticisi tarafından çalıştırılan bir webhook sunucusuna web kancaları kaydeder.
func Example() {
	// Bir yönetici oluşturun
	// Not: GetConfigOrDie, kube-config bulunamazsa herhangi bir mesaj olmadan os.Exit(1) yapacaktır
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		panic(err)
	}

	// Bir webhook sunucusu oluşturun.
	hookServer := NewServer(Options{
		Port: 8443,
	})
	if err := mgr.Add(hookServer); err != nil {
		panic(err)
	}

	// Web kancalarını sunucuya kaydedin.
	hookServer.Register("/mutating", mutatingHook)
	hookServer.Register("/validating", validatingHook)

	// Daha önce ayarlanmış bir yöneticiyi başlatarak sunucuyu başlatın
	err = mgr.Start(ctrl.SetupSignalHandler())
	if err != nil {
		// hatayı işleyin
		panic(err)
	}
}

// Bu örnek, bir denetleyici yöneticisi olmadan çalıştırılabilen bir webhook sunucusu oluşturur.
//
// Bu, varsayılan konumlarda geçerli bir TLS sertifikası ve anahtarı gerektirir
// tls.crt ve tls.key.
func ExampleServer_Start() {
	// Bir webhook sunucusu oluşturun
	hookServer := NewServer(Options{
		Port: 8443,
	})

	// Web kancalarını sunucuya kaydedin.
	hookServer.Register("/mutating", mutatingHook)
	hookServer.Register("/validating", validatingHook)

	// Bir yönetici olmadan sunucuyu başlatın
	err := hookServer.Start(signals.SetupSignalHandler())
	if err != nil {
		// hatayı işleyin
		panic(err)
	}
}

// Bu örnek, bağımsız bir webhook işleyicisi oluşturur
// ve bir denetleyici yöneticisi olmadan bir webhook'u
// mevcut bir sunucuda nasıl çalıştırabileceğinizi göstermek için
// vanilya go HTTP sunucusunda çalıştırır.
func ExampleStandaloneWebhook() {
	// Mevcut bir HTTP sunucunuz olduğunu varsayın
	// istenildiği gibi yapılandırılmış (örneğin TLS ile).
	// Bu örnek için sadece temel bir http.ServeMux oluşturun
	mux := http.NewServeMux()
	port := ":8000"

	// Web kancalarımızdan bağımsız HTTP işleyicileri oluşturun
	mutatingHookHandler, err := admission.StandaloneWebhook(mutatingHook, admission.StandaloneOptions{
		// Logger, isteğe bağlı olarak özel bir logger geçirmenize izin verir
		// (varsayılan olarak log.Log global Logger)
		Logger: logf.RuntimeLog.WithName("mutating-webhook"),
		// MetricsPath, isteğe bağlı olarak
		// prometheus metrikleri için etiketleme amacıyla kullanılacak
		// yolunu sağlar
		// Eğer ayarlanmazsa, prometheus metrikleri oluşturulmaz.
		MetricsPath: "/mutating",
	})
	if err != nil {
		// hatayı işleyin
		panic(err)
	}

	validatingHookHandler, err := admission.StandaloneWebhook(validatingHook, admission.StandaloneOptions{
		Logger:      logf.RuntimeLog.WithName("validating-webhook"),
		MetricsPath: "/validating",
	})
	if err != nil {
		// hatayı işleyin
		panic(err)
	}

	// Webhook işleyicilerini sunucunuza kaydedin
	mux.Handle("/mutating", mutatingHookHandler)
	mux.Handle("/validating", validatingHookHandler)

	// İşleyicinizi çalıştırın
	if err := http.ListenAndServe(port, mux); err != nil { //nolint:gosec // burada zaman aşımı ayarlamamak sorun değil
		panic(err)
	}
}
