/*
2021 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisansa uygun şekilde kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VERİLMEKSİZİN; açık veya zımni garantiler dahil.
Lisans kapsamındaki izinler ve sınırlamalar için Lisansa bakınız.
*/

package internal

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
)

// Selector, ListOptions'a doldurulacak etiket/alan seçicisini belirtir.
type Selector struct {
	Label labels.Selector
	Field fields.Selector
}

// ApplyToList, gerekirse ListOptions'un LabelSelector ve FieldSelector'ını doldurur.
func (s Selector) ApplyToList(listOpts *metav1.ListOptions) {
	if s.Label != nil {
		listOpts.LabelSelector = s.Label.String()
	}
	if s.Field != nil {
		listOpts.FieldSelector = s.Field.String()
	}
}
