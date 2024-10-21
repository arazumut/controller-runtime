/*
2021 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni olarak.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisansa bakın.
*/

package controllertest

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ runtime.Object = &UnconventionalListType{}
var _ runtime.Object = &UnconventionalListTypeList{}

// UnconventionalListType, dilim (slice) türlerinin
// literal dilimlerinden ziyade işaretçi dilimleri olduğu CRD'leri test etmek için kullanılır.
type UnconventionalListType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              string `json:"spec,omitempty"`
}

// DeepCopyObject, runtime.Object'u uygular
// Basitlik için elle yazılmıştır.
func (u *UnconventionalListType) DeepCopyObject() runtime.Object {
	return u.DeepCopy()
}

// DeepCopy, *UnconventionalListType'ı uygular
// Basitlik için elle yazılmıştır.
func (u *UnconventionalListType) DeepCopy() *UnconventionalListType {
	return &UnconventionalListType{
		TypeMeta:   u.TypeMeta,
		ObjectMeta: *u.ObjectMeta.DeepCopy(),
		Spec:       u.Spec,
	}
}

// UnconventionalListTypeList, dilim (slice) türlerinin
// literal dilimlerinden ziyade işaretçi dilimleri olduğu CRD'leri test etmek için kullanılır.
type UnconventionalListTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []*UnconventionalListType `json:"items"`
}

// DeepCopyObject, runtime.Object'u uygular
// Basitlik için elle yazılmıştır.
func (u *UnconventionalListTypeList) DeepCopyObject() runtime.Object {
	return u.DeepCopy()
}

// DeepCopy, *UnconventionalListTypeList'i uygular
// Basitlik için elle yazılmıştır.
func (u *UnconventionalListTypeList) DeepCopy() *UnconventionalListTypeList {
	out := &UnconventionalListTypeList{
		TypeMeta: u.TypeMeta,
		ListMeta: *u.ListMeta.DeepCopy(),
	}
	for _, item := range u.Items {
		out.Items = append(out.Items, item.DeepCopy())
	}
	return out
}
