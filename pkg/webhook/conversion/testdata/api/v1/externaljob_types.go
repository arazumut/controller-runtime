/*

Apache License, Version 2.0 ("Lisans") altında lisanslanmıştır;
bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
bu yazılım Lisans kapsamında "OLDUĞU GİBİ" dağıtılmaktadır,
herhangi bir garanti veya koşul olmaksızın.
Lisans kapsamındaki izin ve sınırlamalar hakkında daha fazla bilgi için
Lisans'a bakınız.
*/

package v1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	v2 "sigs.k8s.io/controller-runtime/pkg/webhook/conversion/testdata/api/v2"
)

// Bu dosyayı düzenleyin! Bu, sizin sahip olmanız için oluşturulmuş bir iskelettir!
// NOT: json etiketleri gereklidir. Eklediğiniz yeni alanların serileştirilmesi için json etiketlerine sahip olması gerekir.

// ExternalJobSpec, ExternalJob'un istenen durumunu tanımlar
type ExternalJobSpec struct {
	// EKSTRA ÖZELLİK ALANLARI EKLEYİN - kümenin istenen durumu
	// Önemli: Bu dosyayı değiştirdikten sonra kodu yeniden oluşturmak için "make" komutunu çalıştırın
	RunAt string `json:"runAt"`
}

// ExternalJobStatus, ExternalJob'un gözlemlenen durumunu tanımlar
type ExternalJobStatus struct {
	// EKSTRA DURUM ALANI EKLEYİN - kümenin gözlemlenen durumu
	// Önemli: Bu dosyayı değiştirdikten sonra kodu yeniden oluşturmak için "make" komutunu çalıştırın
}

// +kubebuilder:object:root=true

// ExternalJob, externaljobs API'si için Şema'dır
type ExternalJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExternalJobSpec   `json:"spec,omitempty"`
	Status ExternalJobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ExternalJobList, bir dizi ExternalJob içerir
type ExternalJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExternalJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ExternalJob{}, &ExternalJobList{})
}

// ConvertTo, Hub türüne (bu durumda v2.ExternalJob) dönüştürme mantığını uygular
func (ej *ExternalJob) ConvertTo(dst conversion.Hub) error {
	switch t := dst.(type) {
	case *v2.ExternalJob:
		jobv2 := dst.(*v2.ExternalJob)
		jobv2.ObjectMeta = ej.ObjectMeta
		jobv2.Spec.ScheduleAt = ej.Spec.RunAt
		return nil
	default:
		return fmt.Errorf("desteklenmeyen tür %v", t)
	}
}

// ConvertFrom, Hub türünden (bu durumda v2.ExternalJob) dönüştürme mantığını uygular
func (ej *ExternalJob) ConvertFrom(src conversion.Hub) error {
	switch t := src.(type) {
	case *v2.ExternalJob:
		jobv2 := src.(*v2.ExternalJob)
		ej.ObjectMeta = jobv2.ObjectMeta
		ej.Spec.RunAt = jobv2.Spec.ScheduleAt
		return nil
	default:
		return fmt.Errorf("desteklenmeyen tür %v", t)
	}
}
