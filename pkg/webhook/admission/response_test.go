/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa veya yazılı izin gereği olmadıkça,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN.
Lisans kapsamındaki izinleri ve sınırlamaları yöneten özel dil için
Lisans'a bakınız.
*/

package admission

import (
	"errors"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	jsonpatch "gomodules.xyz/jsonpatch/v2"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Admission Webhook Yanıt Yardımcıları", func() {
	Describe("Allowed", func() {
		It("izin verilen bir yanıt döndürmelidir", func() {
			Expect(Allowed("")).To(Equal(
				Response{
					AdmissionResponse: admissionv1.AdmissionResponse{
						Allowed: true,
						Result: &metav1.Status{
							Code: http.StatusOK,
						},
					},
				},
			))
		})

		It("bir neden verildiğinde durumu bir neden ile doldurmalıdır", func() {
			Expect(Allowed("kabul edilebilir")).To(Equal(
				Response{
					AdmissionResponse: admissionv1.AdmissionResponse{
						Allowed: true,
						Result: &metav1.Status{
							Code:    http.StatusOK,
							Message: "kabul edilebilir",
						},
					},
				},
			))
		})
	})

	Describe("Denied", func() {
		It("izin verilmeyen bir yanıt döndürmelidir", func() {
			Expect(Denied("")).To(Equal(
				Response{
					AdmissionResponse: admissionv1.AdmissionResponse{
						Allowed: false,
						Result: &metav1.Status{
							Code:   http.StatusForbidden,
							Reason: metav1.StatusReasonForbidden,
						},
					},
				},
			))
		})

		It("bir neden verildiğinde durumu bir neden ile doldurmalıdır", func() {
			Expect(Denied("KABUL EDİLEMEZ!")).To(Equal(
				Response{
					AdmissionResponse: admissionv1.AdmissionResponse{
						Allowed: false,
						Result: &metav1.Status{
							Code:    http.StatusForbidden,
							Reason:  metav1.StatusReasonForbidden,
							Message: "KABUL EDİLEMEZ!",
						},
					},
				},
			))
		})
	})

	Describe("Patched", func() {
		ops := []jsonpatch.JsonPatchOperation{
			{
				Operation: "replace",
				Path:      "/spec/selector/matchLabels",
				Value:     map[string]string{"foo": "bar"},
			},
			{
				Operation: "delete",
				Path:      "/spec/replicas",
			},
		}
		It("verilen yamalarla izin verilen bir yanıt döndürmelidir", func() {
			Expect(Patched("", ops...)).To(Equal(
				Response{
					AdmissionResponse: admissionv1.AdmissionResponse{
						Allowed: true,
						Result: &metav1.Status{
							Code: http.StatusOK,
						},
					},
					Patches: ops,
				},
			))
		})
		It("bir neden verildiğinde durumu bir neden ile doldurmalıdır", func() {
			Expect(Patched("bazı değişiklikler", ops...)).To(Equal(
				Response{
					AdmissionResponse: admissionv1.AdmissionResponse{
						Allowed: true,
						Result: &metav1.Status{
							Code:    http.StatusOK,
							Message: "bazı değişiklikler",
						},
					},
					Patches: ops,
				},
			))
		})
	})

	Describe("Errored", func() {
		It("bir hata ile reddedilen bir yanıt döndürmelidir", func() {
			err := errors.New("bu bir hata")
			expected := Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					Allowed: false,
					Result: &metav1.Status{
						Code:    http.StatusBadRequest,
						Message: err.Error(),
					},
				},
			}
			resp := Errored(http.StatusBadRequest, err)
			Expect(resp).To(Equal(expected))
		})
	})

	Describe("ValidationResponse", func() {
		It("bir mesaj verildiğinde durumu bir mesaj ile doldurmalıdır", func() {
			By("izin verilen yanıtlar için bir mesajın doldurulduğunu kontrol etmek")
			Expect(ValidationResponse(true, "kabul edilebilir")).To(Equal(
				Response{
					AdmissionResponse: admissionv1.AdmissionResponse{
						Allowed: true,
						Result: &metav1.Status{
							Code:    http.StatusOK,
							Message: "kabul edilebilir",
						},
					},
				},
			))

			By("izin verilmeyen yanıtlar için bir mesajın doldurulduğunu kontrol etmek")
			Expect(ValidationResponse(false, "KABUL EDİLEMEZ!")).To(Equal(
				Response{
					AdmissionResponse: admissionv1.AdmissionResponse{
						Allowed: false,
						Result: &metav1.Status{
							Code:    http.StatusForbidden,
							Reason:  metav1.StatusReasonForbidden,
							Message: "KABUL EDİLEMEZ!",
						},
					},
				},
			))
		})

		It("bir kabul kararı döndürmelidir", func() {
			By("izin verildiğinde 'izin verilen' bir yanıt döndürdüğünü kontrol etmek")
			Expect(ValidationResponse(true, "")).To(Equal(
				Response{
					AdmissionResponse: admissionv1.AdmissionResponse{
						Allowed: true,
						Result: &metav1.Status{
							Code: http.StatusOK,
						},
					},
				},
			))

			By("izin verilmediğinde 'izin verilmeyen' bir yanıt döndürdüğünü kontrol etmek")
			Expect(ValidationResponse(false, "")).To(Equal(
				Response{
					AdmissionResponse: admissionv1.AdmissionResponse{
						Allowed: false,
						Result: &metav1.Status{
							Code:   http.StatusForbidden,
							Reason: metav1.StatusReasonForbidden,
						},
					},
				},
			))
		})
	})

	Describe("PatchResponseFromRaw", func() {
		It("iki set seri hale getirilmiş JSON arasındaki farkın yaması ile izin verilen bir yanıt döndürmelidir", func() {
			expected := Response{
				Patches: []jsonpatch.JsonPatchOperation{
					{Operation: "replace", Path: "/a", Value: "bar"},
				},
				AdmissionResponse: admissionv1.AdmissionResponse{
					Allowed:   true,
					PatchType: func() *admissionv1.PatchType { pt := admissionv1.PatchTypeJSONPatch; return &pt }(),
				},
			}
			resp := PatchResponseFromRaw([]byte(`{"a": "foo"}`), []byte(`{"a": "bar"}`))
			Expect(resp).To(Equal(expected))
		})
	})

	Describe("WithWarnings", func() {
		It("mevcut uyarıları kaldırmadan uyarıları mevcut yanıta eklemelidir", func() {
			initialResponse := Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					Allowed: true,
					Result: &metav1.Status{
						Code: http.StatusOK,
					},
					Warnings: []string{"mevcut-uyarı"},
				},
			}
			warnings := []string{"ek-uyarı-1", "ek-uyarı-2"}
			expectedResponse := Response{
				AdmissionResponse: admissionv1.AdmissionResponse{
					Allowed: true,
					Result: &metav1.Status{
						Code: http.StatusOK,
					},
					Warnings: []string{"mevcut-uyarı", "ek-uyarı-1", "ek-uyarı-2"},
				},
			}

			Expect(initialResponse.WithWarnings(warnings...)).To(Equal(expectedResponse))
		})
	})
})
