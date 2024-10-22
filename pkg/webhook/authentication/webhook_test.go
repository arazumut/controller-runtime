/*
2021 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamındaki izinler ve sınırlamalar hakkında daha fazla bilgi için
Lisans'a bakınız.
*/

package authentication

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	machinerytypes "k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Kimlik Doğrulama Webhook'ları", func() {
	allowHandler := func() *Webhook {
		handler := &fakeHandler{
			fn: func(ctx context.Context, req Request) Response {
				return Response{
					TokenReview: authenticationv1.TokenReview{
						Status: authenticationv1.TokenReviewStatus{
							Authenticated: true,
						},
					},
				}
			},
		}
		webhook := &Webhook{
			Handler: handler,
		}

		return webhook
	}

	It("yanıt almak için handler'ı çağırmalı", func() {
		By("izin veren bir handler ile bir webhook kurmak")
		webhook := allowHandler()

		By("webhook'u çağırmak")
		resp := webhook.Handle(context.Background(), Request{})

		By("isteğin izin verildiğini kontrol etmek")
		Expect(resp.Status.Authenticated).To(BeTrue())
	})

	It("yanıtın UID'sinin isteğin UID'sine ayarlandığından emin olmalı", func() {
		By("bir webhook kurmak")
		webhook := allowHandler()

		By("webhook'u çağırmak")
		resp := webhook.Handle(context.Background(), Request{TokenReview: authenticationv1.TokenReview{ObjectMeta: metav1.ObjectMeta{UID: "foobar"}}})

		By("yanıtın isteğin UID'sini paylaştığını kontrol etmek")
		Expect(resp.UID).To(Equal(machinerytypes.UID("foobar")))
	})

	It("yanıtta bir durum sağlanmadıysa durumu doldurmalı", func() {
		By("bir webhook kurmak")
		webhook := allowHandler()

		By("webhook'u çağırmak")
		resp := webhook.Handle(context.Background(), Request{})

		By("yanıtın isteğin UID'sini paylaştığını kontrol etmek")
		Expect(resp.Status).To(Equal(authenticationv1.TokenReviewStatus{Authenticated: true}))
	})

	It("yanıttaki durumu geçersiz kılmamalı", func() {
		By("durum belirten bir webhook kurmak")
		webhook := &Webhook{
			Handler: HandlerFunc(func(ctx context.Context, req Request) Response {
				return Response{
					TokenReview: authenticationv1.TokenReview{
						Status: authenticationv1.TokenReviewStatus{
							Authenticated: true,
							Error:         "Ground Control to Major Tom",
						},
					},
				}
			}),
		}

		By("webhook'u çağırmak")
		resp := webhook.Handle(context.Background(), Request{})

		By("mesajın bozulmadığını kontrol etmek")
		Expect(resp.Status).NotTo(BeNil())
		Expect(resp.Status.Authenticated).To(BeTrue())
		Expect(resp.Status.Error).To(Equal("Ground Control to Major Tom"))
	})
})
