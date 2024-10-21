/*
2021 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMADAN, açık veya zımni.
Lisans kapsamındaki izin ve sınırlamalar için Lisansa bakınız.
*/

package client

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BenchmarkMergeFrom(b *testing.B) {
	cm1 := &corev1.ConfigMap{}
	cm1.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("ConfigMap"))
	cm1.ResourceVersion = "herhangi"

	cm2 := cm1.DeepCopy()
	cm2.Data = map[string]string{"anahtar": "değer"}

	sts1 := &appsv1.StatefulSet{}
	sts1.SetGroupVersionKind(appsv1.SchemeGroupVersion.WithKind("StatefulSet"))
	sts1.ResourceVersion = "birşeyler"

	sts2 := sts1.DeepCopy()
	sts2.Spec.Template.Spec.Containers = []corev1.Container{{
		Resources: corev1.ResourceRequirements{
			Requests: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceCPU:    resource.MustParse("1m"),
				corev1.ResourceMemory: resource.MustParse("1M"),
			},
		},
		ReadinessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{},
			},
		},
		Lifecycle: &corev1.Lifecycle{
			PreStop: &corev1.LifecycleHandler{
				HTTPGet: &corev1.HTTPGetAction{},
			},
		},
		SecurityContext: &corev1.SecurityContext{},
	}}

	b.Run("Seçeneksiz", func(b *testing.B) {
		cmPatch := MergeFrom(cm1)
		if _, err := cmPatch.Data(cm2); err != nil {
			b.Fatalf("beklenen hata yok, alınan %v", err)
		}

		stsPatch := MergeFrom(sts1)
		if _, err := stsPatch.Data(sts2); err != nil {
			b.Fatalf("beklenen hata yok, alınan %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = cmPatch.Data(cm2)
			_, _ = stsPatch.Data(sts2)
		}
	})

	b.Run("İyimserKilitİle", func(b *testing.B) {
		cmPatch := MergeFromWithOptions(cm1, MergeFromWithOptimisticLock{})
		if _, err := cmPatch.Data(cm2); err != nil {
			b.Fatalf("beklenen hata yok, alınan %v", err)
		}

		stsPatch := MergeFromWithOptions(sts1, MergeFromWithOptimisticLock{})
		if _, err := stsPatch.Data(sts2); err != nil {
			b.Fatalf("beklenen hata yok, alınan %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = cmPatch.Data(cm2)
			_, _ = stsPatch.Data(sts2)
		}
	})
}

var _ = Describe("MergeFrom", func() {
	It("iki büyük ve benzer int64 için başarılı bir yama oluşturmalı", func() {
		var büyükInt64 int64 = 9223372036854775807
		var benzerBüyükInt64 int64 = 9223372036854775800
		j := batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      "test",
			},
			Spec: batchv1.JobSpec{
				ActiveDeadlineSeconds: &büyükInt64,
			},
		}
		yama := MergeFrom(j.DeepCopy())

		j.Spec.ActiveDeadlineSeconds = &benzerBüyükInt64

		data, err := yama.Data(&j)
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(Equal([]byte(`{"spec":{"activeDeadlineSeconds":9223372036854775800}}`)))
	})
})
