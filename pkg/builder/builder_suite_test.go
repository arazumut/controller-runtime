/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VERİLMEKSİZİN, açık veya zımni olarak.
Lisans kapsamındaki izinler ve sınırlamalar hakkında daha fazla bilgi için
Lisans'a bakınız.
*/

package builder

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/internal/testing/addr"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// TestBuilder fonksiyonu testleri çalıştırır
func TestBuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Uygulama Suite")
}

var testenv *envtest.Environment
var cfg *rest.Config

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	testenv = &envtest.Environment{}
	addCRDToEnvironment(testenv,
		testDefaulterGVK,
		testValidatorGVK,
		testDefaultValidatorGVK)

	var err error
	cfg, err = testenv.Start()
	Expect(err).NotTo(HaveOccurred())

	// Metrics dinleyicisinin oluşturulmasını engelle
	metricsserver.DefaultBindAddress = "0"

	webhook.DefaultPort, _, err = addr.Suggest("")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	Expect(testenv.Stop()).To(Succeed())

	// DefaultBindAddress'i eski haline getir
	metricsserver.DefaultBindAddress = ":8080"

	// webhook.DefaultPort'u orijinal varsayılan değere geri döndür.
	webhook.DefaultPort = 9443
})

// addCRDToEnvironment fonksiyonu, belirtilen GroupVersionKind'leri test ortamına ekler
func addCRDToEnvironment(env *envtest.Environment, gvks ...schema.GroupVersionKind) {
	for _, gvk := range gvks {
		plural, singular := meta.UnsafeGuessKindToResource(gvk)
		crd := &apiextensionsv1.CustomResourceDefinition{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "apiextensions.k8s.io/v1",
				Kind:       "CustomResourceDefinition",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: plural.Resource + "." + gvk.Group,
			},
			Spec: apiextensionsv1.CustomResourceDefinitionSpec{
				Group: gvk.Group,
				Names: apiextensionsv1.CustomResourceDefinitionNames{
					Plural:   plural.Resource,
					Singular: singular.Resource,
					Kind:     gvk.Kind,
				},
				Scope: apiextensionsv1.NamespaceScoped,
				Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
					{
						Name:    gvk.Version,
						Served:  true,
						Storage: true,
						Schema: &apiextensionsv1.CustomResourceValidation{
							OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
								Type: "object",
							},
						},
					},
				},
			},
		}
		env.CRDInstallOptions.CRDs = append(env.CRDInstallOptions.CRDs, crd)
	}
}
