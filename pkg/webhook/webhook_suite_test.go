/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
yetkiler ve sınırlamalar için Lisans'a bakınız.
*/

package webhook_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// TestSource fonksiyonu, testleri çalıştırmak için kullanılır.
func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Webhook Entegrasyon Test Paketi")
}

var testenv *envtest.Environment
var cfg *rest.Config

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	testenv = &envtest.Environment{}
	// Webhook'u burada başlatıyoruz ve webhook.go'da değil, böylece envtest kurulum kodunu da WebhookOptions ile test edebiliriz.
	initializeWebhookInEnvironment()
	var err error
	cfg, err = testenv.Start()
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	fmt.Println("Durduruluyor?")
	Expect(testenv.Stop()).To(Succeed())
})

// initializeWebhookInEnvironment fonksiyonu, webhook'u test ortamında başlatır.
func initializeWebhookInEnvironment() {
	namespacedScopeV1 := admissionv1.NamespacedScope
	failedTypeV1 := admissionv1.Fail
	equivalentTypeV1 := admissionv1.Equivalent
	noSideEffectsV1 := admissionv1.SideEffectClassNone
	webhookPathV1 := "/failing"

	testenv.WebhookInstallOptions = envtest.WebhookInstallOptions{
		ValidatingWebhooks: []*admissionv1.ValidatingWebhookConfiguration{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "deployment-validation-webhook-config",
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "ValidatingWebhookConfiguration",
					APIVersion: "admissionregistration.k8s.io/v1",
				},
				Webhooks: []admissionv1.ValidatingWebhook{
					{
						Name: "deployment-validation.kubebuilder.io",
						Rules: []admissionv1.RuleWithOperations{
							{
								Operations: []admissionv1.OperationType{"CREATE", "UPDATE"},
								Rule: admissionv1.Rule{
									APIGroups:   []string{"apps"},
									APIVersions: []string{"v1"},
									Resources:   []string{"deployments"},
									Scope:       &namespacedScopeV1,
								},
							},
						},
						FailurePolicy: &failedTypeV1,
						MatchPolicy:   &equivalentTypeV1,
						SideEffects:   &noSideEffectsV1,
						ClientConfig: admissionv1.WebhookClientConfig{
							Service: &admissionv1.ServiceReference{
								Name:      "deployment-validation-service",
								Namespace: "default",
								Path:      &webhookPathV1,
							},
						},
						AdmissionReviewVersions: []string{"v1"},
					},
				},
			},
		},
	}
}
