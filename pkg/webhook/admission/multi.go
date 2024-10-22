/*
Kubernetes Yazarları 2018.

Apache Lisansı, Sürüm 2.0 (Lisans) uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
herhangi bir garanti veya koşul olmaksızın, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
haklar ve sınırlamalar için Lisansa bakınız.
*/

package admission

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	jsonpatch "gomodules.xyz/jsonpatch/v2"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// multiMutating birden fazla Handler'ı içeren bir dilimdir.
type multiMutating []Handler

// Handle, gelen istekleri işleyerek JSON yaması oluşturur.
func (hs multiMutating) Handle(ctx context.Context, req Request) Response {
	patches := []jsonpatch.JsonPatchOperation{}
	for _, handler := range hs {
		resp := handler.Handle(ctx, req)
		if !resp.Allowed {
			return resp
		}
		if resp.PatchType != nil && *resp.PatchType != admissionv1.PatchTypeJSONPatch {
			return Errored(http.StatusInternalServerError,
				fmt.Errorf("beklenmeyen yama türü döndürüldü: %v, yalnızca izin verilen: %v",
					resp.PatchType, admissionv1.PatchTypeJSONPatch))
		}
		patches = append(patches, resp.Patches...)
	}
	var err error
	marshaledPatch, err := json.Marshal(patches)
	if err != nil {
		return Errored(http.StatusBadRequest, fmt.Errorf("yama serileştirilirken hata oluştu: %w", err))
	}
	return Response{
		AdmissionResponse: admissionv1.AdmissionResponse{
			Allowed: true,
			Result: &metav1.Status{
				Code: http.StatusOK,
			},
			Patch:     marshaledPatch,
			PatchType: func() *admissionv1.PatchType { pt := admissionv1.PatchTypeJSONPatch; return &pt }(),
		},
	}
}

// MultiMutatingHandler, birden fazla mutating webhook handler'ını tek bir handler'da birleştirir.
// Handler'lar sıralı olarak çağrılır ve ilk `allowed: false` yanıtı geri kalanını kısa devre yapabilir.
// Kullanıcılar yamaların örtüşmediğinden emin olmalıdır.
func MultiMutatingHandler(handlers ...Handler) Handler {
	return multiMutating(handlers)
}

// multiValidating birden fazla Handler'ı içeren bir dilimdir.
type multiValidating []Handler

// Handle, gelen istekleri işleyerek doğrulama yapar.
func (hs multiValidating) Handle(ctx context.Context, req Request) Response {
	for _, handler := range hs {
		resp := handler.Handle(ctx, req)
		if !resp.Allowed {
			return resp
		}
	}
	return Response{
		AdmissionResponse: admissionv1.AdmissionResponse{
			Allowed: true,
			Result: &metav1.Status{
				Code: http.StatusOK,
			},
		},
	}
}

// MultiValidatingHandler, birden fazla doğrulama webhook handler'ını tek bir handler'da birleştirir.
// Handler'lar sıralı olarak çağrılır ve ilk `allowed: false` yanıtı geri kalanını kısa devre yapabilir.
func MultiValidatingHandler(handlers ...Handler) Handler {
	return multiValidating(handlers)
}
