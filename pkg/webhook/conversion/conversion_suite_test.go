/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") kapsamında lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisansa bakınız.
*/
package conversion

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// TestConversionWebhook, CRD dönüşüm testlerini çalıştırır
func TestConversionWebhook(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CRD Dönüşüm Test Paketi")
}

// BeforeSuite, testler başlamadan önce çalıştırılacak olan ayarları yapar
var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
})
