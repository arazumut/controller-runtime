/*

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni olarak.
Lisans kapsamında izin verilen belirli dil kapsamındaki
yetkiler ve sınırlamalar için Lisansa bakınız.
*/

// Paket v1, jobs v1 API grubu için API Şeması tanımlarını içerir
// +kubebuilder:object:generate=true
// +groupName=jobs.testprojects.kb.io
package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion, bu nesneleri kaydetmek için kullanılan grup sürümüdür
	GroupVersion = schema.GroupVersion{Group: "jobs.testprojects.kb.io", Version: "v1"}

	// SchemeBuilder, Go türlerini GroupVersionKind şemasına eklemek için kullanılır
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme, bu grup-sürümdeki türleri verilen şemaya ekler.
	AddToScheme = SchemeBuilder.AddToScheme
)
