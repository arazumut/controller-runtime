/*

Apache License, Version 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni olarak.
Lisans kapsamındaki izin ve sınırlamalar hakkında daha fazla bilgi için
Lisans'a bakınız.
*/

package main

import (
	"context"
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	jobsv1 "sigs.k8s.io/controller-runtime/pkg/webhook/conversion/testdata/api/v1"
	jobsv2 "sigs.k8s.io/controller-runtime/pkg/webhook/conversion/testdata/api/v2"
	jobsv3 "sigs.k8s.io/controller-runtime/pkg/webhook/conversion/testdata/api/v3"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	// API şemalarını ekliyoruz
	jobsv1.AddToScheme(scheme)
	jobsv2.AddToScheme(scheme)
	jobsv3.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "Metriğin bağlanacağı adres.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Lider seçim özelliğini etkinleştir. Bu, yalnızca bir aktif kontrol yöneticisi olmasını sağlar.")
	flag.Parse()

	ctrl.SetLogger(zap.Logger(true))

	mgr, err := ctrl.NewManager(context.Background(), ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:         scheme,
		Metrics:        metricsserver.Options{BindAddress: metricsAddr},
		LeaderElection: enableLeaderElection,
	})
	if err != nil {
		setupLog.Error(err, "yönetici başlatılamadı")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("yönetici başlatılıyor")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "yönetici çalıştırılırken sorun oluştu")
		os.Exit(1)
	}
}
