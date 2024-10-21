/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisansa bakınız.
*/

package main

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-v1-pod,mutating=false,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=vpod.kb.io

// podValidator, Pod'ları oluşturma ve güncelleme işlemleri sırasında doğrular
type podValidator struct{}

// validate, belirli bir anotasyonun var olup olmadığını ve doğru değere sahip olup olmadığını kontrol eder
func (v *podValidator) validate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	log := logf.FromContext(ctx)

	// obj'nin bir Pod olduğundan emin olmak için tür dönüşümü
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return nil, fmt.Errorf("Pod bekleniyordu ama %T alındı", obj)
	}

	log.Info("Pod doğrulanıyor", "podName", pod.Name)

	// Belirli bir anotasyonu kontrol et
	key := "example-mutating-admission-webhook"
	anno, found := pod.Annotations[key]
	if !found {
		log.Info("Pod gerekli anotasyonu içermiyor", "annotationKey", key)
		return nil, fmt.Errorf("gerekli anotasyon eksik: %s", key)
	}

	// Anotasyon değerini doğrula
	if anno != "foo" {
		log.Info("Pod yanlış anotasyon değerine sahip", "expected", "foo", "found", anno)
		return nil, fmt.Errorf("anotasyon %s beklenen değere sahip değil %q", key, "foo")
	}

	log.Info("Pod doğrulamayı geçti", "podName", pod.Name)
	return nil, nil
}

// ValidateCreate, Pod oluşturma sırasında çağrılır
func (v *podValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	log := logf.FromContext(ctx)
	log.Info("Pod oluşturma doğrulanıyor")
	return v.validate(ctx, obj)
}

// ValidateUpdate, Pod güncelleme sırasında çağrılır
func (v *podValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	log := logf.FromContext(ctx)
	log.Info("Pod güncelleme doğrulanıyor")
	return v.validate(ctx, newObj)
}
