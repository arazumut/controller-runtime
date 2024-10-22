/*

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
haklar ve sınırlamalar için Lisansa bakınız.
*/

package v2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BU DOSYAYI DÜZENLEYİN! BU, SAHİP OLACAĞINIZ BİR İSKELETTİR!
// NOT: json etiketleri gereklidir. Eklediğiniz yeni alanların serileştirilmesi için json etiketlerine sahip olması gerekir.

// ExternalJobSpec, ExternalJob'un istenen durumunu tanımlar
type ExternalJobSpec struct {
	// EK ÖZELLİK ALANLARI EKLEYİN - kümenin istenen durumu
	// Önemli: Bu dosyayı değiştirdikten sonra kodu yeniden oluşturmak için "make" komutunu çalıştırın
	ScheduleAt string `json:"scheduleAt"`
}

// ExternalJobStatus, ExternalJob'un gözlemlenen durumunu tanımlar
type ExternalJobStatus struct {
	// EK DURUM ALANI EKLEYİN - kümenin gözlemlenen durumu
	// Önemli: Bu dosyayı değiştirdikten sonra kodu yeniden oluşturmak için "make" komutunu çalıştırın
}

// +kubebuilder:object:root=true

// ExternalJob, externaljobs API'si için Şemadır
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

// Hub, v2.ExternalJob'un bu durumda Hub türü olduğunu belirtmek için bir işaretleyici yöntemdir.
// v2.ExternalJob depolama sürümüdür, bu nedenle bunu Hub olarak işaretleyin.
// Depolama sürümü herhangi bir dönüştürme yöntemi uygulamak zorunda değildir çünkü
// varsayılan conversionHandler depolama sürümü için dönüştürme mantığını uygular.
// TODO: Bunu depolama sürümü olarak işaretlemek için buraya yorum ekleyin
func (ej *ExternalJob) Hub() {}
