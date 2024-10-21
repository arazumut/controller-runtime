/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ OLMAKSIZIN, açık veya zımni olarak.
Lisans kapsamındaki izin ve sınırlamalar için Lisansa bakınız.
*/

package builder

import (
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// {{{ "Fonksiyonel" Seçenek Arayüzleri

// ForOption, bir For isteği için seçenekleri değiştiren bazı yapılandırmalardır.
type ForOption interface {
	// ApplyToFor, bu yapılandırmayı verilen for girdisine uygular.
	ApplyToFor(*ForInput)
}

// OwnsOption, bir owns isteği için seçenekleri değiştiren bazı yapılandırmalardır.
type OwnsOption interface {
	// ApplyToOwns, bu yapılandırmayı verilen owns girdisine uygular.
	ApplyToOwns(*OwnsInput)
}

// WatchesOption, bir watches isteği için seçenekleri değiştiren bazı yapılandırmalardır.
type WatchesOption interface {
	// ApplyToWatches, bu yapılandırmayı verilen watches seçeneklerine uygular.
	ApplyToWatches(untypedWatchesInput)
}

// }}}

// {{{ Çoklu Tür Seçenekleri

// WithPredicates, verilen predicate listesini ayarlar.
func WithPredicates(predicates ...predicate.Predicate) Predicates {
	return Predicates{
		predicates: predicates,
	}
}

// Predicates, anahtarları kuyruğa almadan önce olayları filtreler.
type Predicates struct {
	predicates []predicate.Predicate
}

// ApplyToFor, bu yapılandırmayı verilen ForInput seçeneklerine uygular.
func (w Predicates) ApplyToFor(opts *ForInput) {
	opts.predicates = w.predicates
}

// ApplyToOwns, bu yapılandırmayı verilen OwnsInput seçeneklerine uygular.
func (w Predicates) ApplyToOwns(opts *OwnsInput) {
	opts.predicates = w.predicates
}

// ApplyToWatches, bu yapılandırmayı verilen WatchesInput seçeneklerine uygular.
func (w Predicates) ApplyToWatches(opts untypedWatchesInput) {
	opts.setPredicates(w.predicates)
}

var _ ForOption = &Predicates{}
var _ OwnsOption = &Predicates{}
var _ WatchesOption = &Predicates{}

// }}}

// {{{ For & Owns Çift Tür Seçenekleri

// projectAs, girdideki projeksiyonu yapılandırır.
// Şu anda yalnızca OnlyMetadata desteklenmektedir. Gelecekte
// keyfi olmayan yerel projeksiyonları genişletmek isteyebiliriz.
type projectAs objectProjection

// ApplyToFor, bu yapılandırmayı verilen ForInput seçeneklerine uygular.
func (p projectAs) ApplyToFor(opts *ForInput) {
	opts.objectProjection = objectProjection(p)
}

// ApplyToOwns, bu yapılandırmayı verilen OwnsInput seçeneklerine uygular.
func (p projectAs) ApplyToOwns(opts *OwnsInput) {
	opts.objectProjection = objectProjection(p)
}

// ApplyToWatches, bu yapılandırmayı verilen WatchesInput seçeneklerine uygular.
func (p projectAs) ApplyToWatches(opts untypedWatchesInput) {
	opts.setObjectProjection(objectProjection(p))
}

var (
	// OnlyMetadata, denetleyiciye yalnızca meta verileri önbelleğe almasını ve
	// API sunucusunu yalnızca meta veri formunda izlemesini söyler. Bu, birçok
	// nesneyi izlerken, gerçekten büyük nesneler veya yalnızca GVK'yı bildiğiniz
	// ancak yapısını bilmediğiniz nesneler için yararlıdır. Nesneleri
	// reconciler'da alırken istemciye metav1.PartialObjectMetadata geçirmeniz
	// gerekecek, aksi takdirde yapılandırılmış veya yapılandırılmamış önbelleğin
	// bir kopyasını oluşturursunuz.
	//
	// OnlyMetadata ile bir kaynağı izlerken, örneğin v1.Pod,
	// v1.Pod türünü kullanarak Get ve List yapmamalısınız. Bunun yerine,
	// özel metav1.PartialObjectMetadata türünü kullanmalısınız.
	//
	// ❌ Yanlış:
	//
	//   pod := &v1.Pod{}
	//   mgr.GetClient().Get(ctx, nsAndName, pod)
	//
	// ✅ Doğru:
	//
	//   pod := &metav1.PartialObjectMetadata{}
	//   pod.SetGroupVersionKind(schema.GroupVersionKind{
	//       Group:   "",
	//       Version: "v1",
	//       Kind:    "Pod",
	//   })
	//   mgr.GetClient().Get(ctx, nsAndName, pod)
	//
	// İlk durumda, controller-runtime meta veri önbelleğinin üzerine başka bir
	// önbellek oluşturur; bu, bellek tüketimini artırır ve önbellekler senkronize
	// olmadığından yarış koşullarına yol açar.
	OnlyMetadata = projectAs(projectAsMetadata)

	_ ForOption     = OnlyMetadata
	_ OwnsOption    = OnlyMetadata
	_ WatchesOption = OnlyMetadata
)

// }}}

// MatchEveryOwner, izleme işleminin kontrol sahibi sahipliğine göre filtrelenip
// filtrelenmeyeceğini belirler. Yani, OwnerReference.Controller alanı ayarlandığında.
//
// Bir seçenek olarak geçilirse,
// işleyici, verilen türdeki nesnenin her sahibi için bildirim alır.
// Ayarlanmazsa (varsayılan), işleyici yalnızca `Controller: true` olan
// ilk OwnerReference için bildirim alır.
var MatchEveryOwner = &matchEveryOwner{}

type matchEveryOwner struct{}

// ApplyToOwns, bu yapılandırmayı verilen OwnsInput seçeneklerine uygular.
func (o matchEveryOwner) ApplyToOwns(opts *OwnsInput) {
	opts.matchEveryOwner = true
}
