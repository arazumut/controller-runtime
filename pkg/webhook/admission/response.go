/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa uyarınca veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
haklar ve sınırlamalar için Lisansa bakınız.
*/

package admission

import (
	"net/http"

	jsonpatch "gomodules.xyz/jsonpatch/v2"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Allowed, verilen işlemin izinli olduğunu belirten bir yanıt oluşturur (herhangi bir yama olmadan).
func Allowed(message string) Response {
	return ValidationResponse(true, message)
}

// Denied, verilen işlemin izinli olmadığını belirten bir yanıt oluşturur.
func Denied(message string) Response {
	return ValidationResponse(false, message)
}

// Patched, verilen işlemin izinli olduğunu ve hedef nesnenin verilen JSONPatch işlemleriyle değiştirilmesi gerektiğini belirten bir yanıt oluşturur.
func Patched(message string, patches ...jsonpatch.JsonPatchOperation) Response {
	resp := Allowed(message)
	resp.Patches = patches

	return resp
}

// Errored, bir isteği hata ile işlemek için yeni bir Yanıt oluşturur.
func Errored(code int32, err error) Response {
	return Response{
		AdmissionResponse: admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Code:    code,
				Message: err.Error(),
			},
		},
	}
}

// ValidationResponse, bir isteği kabul etmek için bir yanıt döndürür.
func ValidationResponse(allowed bool, message string) Response {
	code := http.StatusForbidden
	reason := metav1.StatusReasonForbidden
	if allowed {
		code = http.StatusOK
		reason = ""
	}
	resp := Response{
		AdmissionResponse: admissionv1.AdmissionResponse{
			Allowed: allowed,
			Result: &metav1.Status{
				Code:   int32(code), //nolint:gosec // Burada tamsayı taşmaları (G115) meydana gelemez.
				Reason: reason,
			},
		},
	}
	if len(message) > 0 {
		resp.Result.Message = message
	}
	return resp
}

// PatchResponseFromRaw, 2 bayt dizisi alır ve json yaması ile yeni bir yanıt döndürür.
// Orijinal nesne, https://github.com/kubernetes-sigs/kubebuilder/issues/510 adresinde açıklanan roundtripping sorununu önlemek için ham bayt olarak geçirilmelidir.
func PatchResponseFromRaw(original, current []byte) Response {
	patches, err := jsonpatch.CreatePatch(original, current)
	if err != nil {
		return Errored(http.StatusInternalServerError, err)
	}
	return Response{
		Patches: patches,
		AdmissionResponse: admissionv1.AdmissionResponse{
			Allowed: true,
			PatchType: func() *admissionv1.PatchType {
				if len(patches) == 0 {
					return nil
				}
				pt := admissionv1.PatchTypeJSONPatch
				return &pt
			}(),
		},
	}
}

// validationResponseFromStatus, sağlanan Durum nesnesi ile bir isteği kabul etmek için bir yanıt döndürür.
func validationResponseFromStatus(allowed bool, status metav1.Status) Response {
	resp := Response{
		AdmissionResponse: admissionv1.AdmissionResponse{
			Allowed: allowed,
			Result:  &status,
		},
	}
	return resp
}

// WithWarnings, verilen uyarıları Yanıt'a ekler.
// Önceden herhangi bir uyarı verilmişse, bunlar üzerine yazılmaz.
func (r Response) WithWarnings(warnings ...string) Response {
	r.AdmissionResponse.Warnings = append(r.AdmissionResponse.Warnings, warnings...)
	return r
}
