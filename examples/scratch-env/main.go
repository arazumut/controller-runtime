/*
Telif Hakkı 2021 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izinle aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisansa bakın.
*/

package main

import (
	goflag "flag"
	"os"

	flag "github.com/spf13/pflag"
	"go.uber.org/zap"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	crdPaths              = flag.StringSlice("crd-paths", nil, "Başlangıçta yüklenecek CRD dosyalarının veya dizinlerinin yolları")
	webhookPaths          = flag.StringSlice("webhook-paths", nil, "Başlangıçta yüklenecek webhook yapılandırma dosyalarının veya dizinlerinin yolları")
	attachControlPlaneOut = flag.Bool("debug-env", false, "Test ortamına (apiserver & etcd) çıktı ekle -- KUBEBUILDER_ATTACH_CONTROL_PLANE_OUTPUT=true zorlamak için bir kolaylık bayrağı")
)

// main, kontrol yöneticisi için giriş noktasıdır.

// defer'leri atlamadan bir çıkış kodu döndürebilmek için ayrı bir fonksiyon
func runMain() int {
	loggerOpts := &logzap.Options{
		Development: true, // mantıklı bir varsayılan
		ZapOpts:     []zap.Option{zap.AddCaller()},
	}
	{
		var goFlagSet goflag.FlagSet
		loggerOpts.BindFlags(&goFlagSet)
		flag.CommandLine.AddGoFlagSet(&goFlagSet)
	}
	flag.Parse()
	ctrl.SetLogger(logzap.New(logzap.UseFlagOptions(loggerOpts)))
	ctrl.Log.Info("Başlatılıyor...")

	log := ctrl.Log.WithName("main")

	env := &envtest.Environment{}
	env.CRDInstallOptions.Paths = *crdPaths
	env.WebhookInstallOptions.Paths = *webhookPaths

	if *attachControlPlaneOut {
		os.Setenv("KUBEBUILDER_ATTACH_CONTROL_PLANE_OUTPUT", "true")
	}

	log.Info("apiserver & etcd başlatılıyor")
	cfg, err := env.Start()
	if err != nil {
		log.Error(err, "test ortamı başlatılamadı")
		// CRD'leri yüklerken veya kullanıcıları sağlarken başarısız olursak ortamı kapatın.
		if err := env.Stop(); err != nil {
			log.Error(err, "hata sonrası test ortamı durdurulamadı (bu beklenebilir, ancak bilmenizi istedik)")
		}
		return 1
	}

	log.Info("apiserver çalışıyor", "host", cfg.Host)

	// NB(directxman12): bu grup maalesef adlandırılmıştır, ancak çeşitli
	// kubernetes sürümleri bize "admin" erişimi sağlamak için bunu kullanmamızı gerektirir.
	user, err := env.ControlPlane.AddUser(envtest.User{
		Name:   "envtest-admin",
		Groups: []string{"system:masters"},
	}, nil)
	if err != nil {
		log.Error(err, "admin kullanıcı sağlanamadı, onsuz devam ediliyor")
		return 1
	}

	// TODO: mevcut bir dosyada yeni bir bağlama yazma desteği ekleyin
	kubeconfigFile, err := os.CreateTemp("", "scratch-env-kubeconfig-")
	if err != nil {
		log.Error(err, "kubeconfig dosyası oluşturulamadı, onsuz devam ediliyor")
		return 1
	}
	defer os.Remove(kubeconfigFile.Name())

	{
		log := log.WithValues("path", kubeconfigFile.Name())
		log.V(1).Info("kubeconfig yazılıyor")

		kubeConfig, err := user.KubeConfig()
		if err != nil {
			log.Error(err, "kubeconfig oluşturulamadı")
		}

		if _, err := kubeconfigFile.Write(kubeConfig); err != nil {
			log.Error(err, "kubeconfig kaydedilemedi")
			return 1
		}

		log.Info("kubeconfig yazıldı")
	}

	if opts := env.WebhookInstallOptions; opts.LocalServingPort != 0 {
		log.Info("webhook'lar yapılandırıldı", "host", opts.LocalServingHost, "port", opts.LocalServingPort, "dir", opts.LocalServingCertDir)
	}

	ctx := ctrl.SetupSignalHandler()
	<-ctx.Done()

	log.Info("apiserver & etcd kapatılıyor")
	err = env.Stop()
	if err != nil {
		log.Error(err, "test ortamı durdurulamadı")
		return 1
	}

	log.Info("Başarıyla kapatıldı")
	return 0
}

func main() {
	os.Exit(runMain())
}
