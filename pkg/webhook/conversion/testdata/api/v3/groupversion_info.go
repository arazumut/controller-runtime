/*

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisans'ın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamındaki izinler ve sınırlamalar hakkında daha fazla bilgi için
Lisans'a bakınız.
*/

// Paket v3, jobs v3 API grubu için API Şeması tanımlarını içerir
// +kubebuilder:object:generate=true
// +groupName=jobs.testprojects.kb.io
package v3

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion, bu nesneleri kaydetmek için kullanılan grup sürümüdür
	GroupVersion = schema.GroupVersion{Group: "jobs.testprojects.kb.io", Version: "v3"}

	// SchemeBuilder, Go türlerini GroupVersionKind şemasına eklemek için kullanılır
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme, bu grup-sürümdeki türleri verilen şemaya ekler.
	AddToScheme = SchemeBuilder.AddToScheme
)
