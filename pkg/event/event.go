/*
2018 Kubernetes Yazarları tarafından oluşturulmuştur.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisans'a bakınız.
*/

package event

import "sigs.k8s.io/controller-runtime/pkg/client"

// CreateEvent, bir Kubernetes nesnesinin oluşturulduğu bir olaydır. CreateEvent, bir source.Source tarafından oluşturulmalı
// ve bir handler.EventHandler tarafından reconcile.Request'e dönüştürülmelidir.
type CreateEvent = TypedCreateEvent[client.Object]

// UpdateEvent, bir Kubernetes nesnesinin güncellendiği bir olaydır. UpdateEvent, bir source.Source tarafından oluşturulmalı
// ve bir handler.EventHandler tarafından reconcile.Request'e dönüştürülmelidir.
type UpdateEvent = TypedUpdateEvent[client.Object]

// DeleteEvent, bir Kubernetes nesnesinin silindiği bir olaydır. DeleteEvent, bir source.Source tarafından oluşturulmalı
// ve bir handler.EventHandler tarafından reconcile.Request'e dönüştürülmelidir.
type DeleteEvent = TypedDeleteEvent[client.Object]

// GenericEvent, işlem türünün bilinmediği bir olaydır (örneğin, küme dışından gelen olaylar veya anketler).
// GenericEvent, bir source.Source tarafından oluşturulmalı ve bir handler.EventHandler tarafından reconcile.Request'e dönüştürülmelidir.
type GenericEvent = TypedGenericEvent[client.Object]

// TypedCreateEvent, bir Kubernetes nesnesinin oluşturulduğu bir olaydır. TypedCreateEvent, bir source.Source tarafından oluşturulmalı
// ve bir handler.TypedEventHandler tarafından reconcile.Request'e dönüştürülmelidir.
type TypedCreateEvent[object any] struct {
	// Object, olaydan gelen nesnedir
	Object object
}

// TypedUpdateEvent, bir Kubernetes nesnesinin güncellendiği bir olaydır. TypedUpdateEvent, bir source.Source tarafından oluşturulmalı
// ve bir handler.TypedEventHandler tarafından reconcile.Request'e dönüştürülmelidir.
type TypedUpdateEvent[object any] struct {
	// ObjectOld, olaydan gelen eski nesnedir
	ObjectOld object

	// ObjectNew, olaydan gelen yeni nesnedir
	ObjectNew object
}

// TypedDeleteEvent, bir Kubernetes nesnesinin silindiği bir olaydır. TypedDeleteEvent, bir source.Source tarafından oluşturulmalı
// ve bir handler.TypedEventHandler tarafından reconcile.Request'e dönüştürülmelidir.
type TypedDeleteEvent[object any] struct {
	// Object, olaydan gelen nesnedir
	Object object

	// DeleteStateUnknown, Silme olayının kaçırıldığını ancak nesnenin silindiğini belirlediğimizi gösterir.
	DeleteStateUnknown bool
}

// TypedGenericEvent, işlem türünün bilinmediği bir olaydır (örneğin, küme dışından gelen olaylar veya anketler).
// TypedGenericEvent, bir source.Source tarafından oluşturulmalı ve bir handler.TypedEventHandler tarafından reconcile.Request'e dönüştürülmelidir.
type TypedGenericEvent[object any] struct {
	// Object, olaydan gelen nesnedir
	Object object
}
