/*
2018 Kubernetes Yazarları tarafından oluşturulmuştur.

Apache Lisansı, Sürüm 2.0 ("Lisans") kapsamında lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
yetkiler ve sınırlamalar için Lisans'a bakınız.
*/

package handler_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// TestEventhandler, testlerin çalıştırılmasını sağlar
func TestEventhandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Eventhandler Suite")
}

var testenv *envtest.Environment
var cfg *rest.Config

// BeforeSuite, testler başlamadan önce çalıştırılır
var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	testenv = &envtest.Environment{}
	var err error
	cfg, err = testenv.Start()
	Expect(err).NotTo(HaveOccurred())
})

// AfterSuite, testler bittikten sonra çalıştırılır
var _ = AfterSuite(func() {
	Expect(testenv.Stop()).To(Succeed())
})
