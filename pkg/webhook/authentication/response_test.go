/*
2021 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izinle gerekli olmadıkça,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
haklar ve sınırlamalar için Lisansa bakın.
*/

package authentication

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	authenticationv1 "k8s.io/api/authentication/v1"
)

var _ = Describe("Kimlik Doğrulama Webhook Yanıt Yardımcıları", func() {
	Describe("Authenticated", func() {
		It("izin verilen bir yanıt döndürmelidir", func() {
			Expect(Authenticated("", authenticationv1.UserInfo{})).To(Equal(
				Response{
					TokenReview: authenticationv1.TokenReview{
						Status: authenticationv1.TokenReviewStatus{
							Authenticated: true,
							User:          authenticationv1.UserInfo{},
						},
					},
				},
			))
		})

		It("bir neden verildiğinde bir durumun neden ile doldurulması gerekir", func() {
			Expect(Authenticated("kabul edilebilir", authenticationv1.UserInfo{})).To(Equal(
				Response{
					TokenReview: authenticationv1.TokenReview{
						Status: authenticationv1.TokenReviewStatus{
							Authenticated: true,
							User:          authenticationv1.UserInfo{},
							Error:         "kabul edilebilir",
						},
					},
				},
			))
		})
	})

	Describe("Unauthenticated", func() {
		It("izin verilmeyen bir yanıt döndürmelidir", func() {
			Expect(Unauthenticated("", authenticationv1.UserInfo{})).To(Equal(
				Response{
					TokenReview: authenticationv1.TokenReview{
						Status: authenticationv1.TokenReviewStatus{
							Authenticated: false,
							User:          authenticationv1.UserInfo{},
							Error:         "",
						},
					},
				},
			))
		})

		It("bir neden verildiğinde bir durumun neden ile doldurulması gerekir", func() {
			Expect(Unauthenticated("KABUL EDİLEMEZ!", authenticationv1.UserInfo{})).To(Equal(
				Response{
					TokenReview: authenticationv1.TokenReview{
						Status: authenticationv1.TokenReviewStatus{
							Authenticated: false,
							User:          authenticationv1.UserInfo{},
							Error:         "KABUL EDİLEMEZ!",
						},
					},
				},
			))
		})
	})

	Describe("Errored", func() {
		It("bir hata ile kimlik doğrulaması yapılmamış bir yanıt döndürmelidir", func() {
			err := errors.New("bu bir hatadır")
			expected := Response{
				TokenReview: authenticationv1.TokenReview{
					Status: authenticationv1.TokenReviewStatus{
						Authenticated: false,
						User:          authenticationv1.UserInfo{},
						Error:         err.Error(),
					},
				},
			}
			resp := Errored(err)
			Expect(resp).To(Equal(expected))
		})
	})

	Describe("ReviewResponse", func() {
		It("bir neden verildiğinde bir durumun hata ile doldurulması gerekir", func() {
			By("izin verilen yanıtlar için bir mesajın doldurulduğunu kontrol etmek")
			Expect(ReviewResponse(true, authenticationv1.UserInfo{}, "kabul edilebilir")).To(Equal(
				Response{
					TokenReview: authenticationv1.TokenReview{
						Status: authenticationv1.TokenReviewStatus{
							Authenticated: true,
							User:          authenticationv1.UserInfo{},
							Error:         "kabul edilebilir",
						},
					},
				},
			))

			By("kimlik doğrulaması yapılmamış yanıtlar için bir mesajın doldurulduğunu kontrol etmek")
			Expect(ReviewResponse(false, authenticationv1.UserInfo{}, "KABUL EDİLEMEZ!")).To(Equal(
				Response{
					TokenReview: authenticationv1.TokenReview{
						Status: authenticationv1.TokenReviewStatus{
							Authenticated: false,
							User:          authenticationv1.UserInfo{},
							Error:         "KABUL EDİLEMEZ!",
						},
					},
				},
			))
		})

		It("bir kimlik doğrulama kararı döndürmelidir", func() {
			By("izin verildiğinde 'izin verilen' bir yanıt döndürdüğünü kontrol etmek")
			Expect(ReviewResponse(true, authenticationv1.UserInfo{}, "")).To(Equal(
				Response{
					TokenReview: authenticationv1.TokenReview{
						Status: authenticationv1.TokenReviewStatus{
							Authenticated: true,
							User:          authenticationv1.UserInfo{},
						},
					},
				},
			))

			By("izin verilmediğinde 'kimlik doğrulaması yapılmamış' bir yanıt döndürdüğünü kontrol etmek")
			Expect(ReviewResponse(false, authenticationv1.UserInfo{}, "")).To(Equal(
				Response{
					TokenReview: authenticationv1.TokenReview{
						Status: authenticationv1.TokenReviewStatus{
							Authenticated: false,
							User:          authenticationv1.UserInfo{},
						},
					},
				},
			))
		})
	})
})
