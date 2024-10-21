/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
herhangi bir garanti veya koşul olmaksızın, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
haklar ve sınırlamalar için Lisans'a bakınız.
*/

package main

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=mpod.kb.io

// podAnnotator, Pod'ları özel meta verilerle anotasyonlamak için kullanılan webhook'tur
type podAnnotator struct{}

// Create, Pod oluşturma mutasyon isteklerini işler
func (a *podAnnotator) Create(ctx context.Context, obj runtime.Object) error {
	return a.Default(ctx, obj) // Varsayılan mutasyon mantığı yeniden kullanılır
}

// Update, Pod güncelleme mutasyon isteklerini işler
func (a *podAnnotator) Update(ctx context.Context, obj runtime.Object) error {
	return a.Default(ctx, obj) // Varsayılan mutasyon mantığı yeniden kullanılır
}

// Delete, Pod silme işlemlerini işler, mutasyonlar için burada yapılacak bir şey yok
func (a *podAnnotator) Delete(ctx context.Context, obj runtime.Object) error {
	return nil
}

// Default, Pod webhook'u için mutasyon mantığını uygular
func (a *podAnnotator) Default(ctx context.Context, obj runtime.Object) error {
	log := logf.FromContext(ctx)

	// Nesneyi bir Pod olarak dönüştürerek tür doğruluğunu sağlar
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return fmt.Errorf("Pod bekleniyordu ama %T alındı", obj)
	}

	// Pod'un anotasyonları olduğundan emin olun
	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}

	// Özel anotasyonu ekleyin veya değiştirin
	pod.Annotations["example-mutating-admission-webhook"] = "foo"
	log.Info("Pod 'example-mutating-admission-webhook' ile anotasyonlandı")

	return nil
}
