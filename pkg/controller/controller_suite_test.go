/*
Telif Hakkı 2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN.
Lisans kapsamında izin verilen belirli dil kapsamındaki
haklar ve sınırlamalar için Lisansa bakın.
*/

package controller_test

import (
	"net/http"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllertest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	crscheme "sigs.k8s.io/controller-runtime/pkg/scheme"
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Entegrasyon Testi")
}

var testenv *envtest.Environment
var cfg *rest.Config
var clientset *kubernetes.Clientset

// clientTransport, sızıntıları kontrol eden testlerde keep-alive bağlantılarını zorla kapatmak için kullanılır.
var clientTransport *http.Transport

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	err := (&crscheme.Builder{
		GroupVersion: schema.GroupVersion{Group: "chaosapps.metamagical.io", Version: "v1"},
	}).
		Register(
			&controllertest.UnconventionalListType{},
			&controllertest.UnconventionalListTypeList{},
		).AddToScheme(scheme.Scheme)
	Expect(err).ToNot(HaveOccurred())

	testenv = &envtest.Environment{
		CRDDirectoryPaths: []string{"testdata/crds"},
	}

	cfg, err = testenv.Start()
	Expect(err).NotTo(HaveOccurred())

	cfg.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
		// NB(directxman12): Transport'u *ve* TLS seçeneklerini kullanamayız,
		// bu yüzden oluşturulduktan hemen sonra transport'u alıyoruz ki
		// üzerinde tür iddiasında bulunabilelim (umarım)?
		// umarım bu kırılmaz 🤞
		clientTransport = rt.(*http.Transport)
		return rt
	}

	clientset, err = kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())

	// Metrics dinleyicisinin oluşturulmasını engelle
	metricsserver.DefaultBindAddress = "0"
})

var _ = AfterSuite(func() {
	Expect(testenv.Stop()).To(Succeed())

	// DefaultBindAddress'i geri yükle
	metricsserver.DefaultBindAddress = ":8080"
})
