/*
2021 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisansa bakın.
*/

package main

import (
	"os"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook/authentication"
)

func init() {
	log.SetLogger(zap.New())
}

func main() {
	girisLog := log.Log.WithName("giriş noktası")

	// Bir Yönetici Ayarlayın
	girisLog.Info("yönetici ayarlanıyor")
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		girisLog.Error(err, "genel kontrol yöneticisi ayarlanamıyor")
		os.Exit(1)
	}

	// Kontrolcülerini buraya ekleyin

	// Webhook'ları Ayarlayın
	girisLog.Info("webhook sunucusu ayarlanıyor")
	hookServer := mgr.GetWebhookServer()

	girisLog.Info("webhook'ları webhook sunucusuna kaydediyor")
	hookServer.Register("/validate-v1-tokenreview", &authentication.Webhook{Handler: &authenticator{}})

	girisLog.Info("yönetici başlatılıyor")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		girisLog.Error(err, "yönetici çalıştırılamıyor")
		os.Exit(1)
	}
	girisLog.Info("yönetici başlatılıyor")
}
