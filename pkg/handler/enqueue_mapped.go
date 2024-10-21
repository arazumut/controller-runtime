/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
haklar ve sınırlamalar için Lisansa bakın.
*/

package handler

import (
	"context"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// MapFunc, genel bir fonksiyondan istekleri kuyruğa almak için gereken imzadır.
// Bu tür genellikle bir olay işleyici kaydederken EnqueueRequestsFromMapFunc ile kullanılır.
type MapFunc = TypedMapFunc[client.Object, reconcile.Request]

// TypedMapFunc, genel bir fonksiyondan istekleri kuyruğa almak için gereken imzadır.
// Bu tür genellikle bir olay işleyici kaydederken EnqueueRequestsFromTypedMapFunc ile kullanılır.
//
// TypedMapFunc deneysel olup gelecekte değişikliğe tabidir.
type TypedMapFunc[object any, request comparable] func(context.Context, object) []request

// EnqueueRequestsFromMapFunc, her Olayda bir dizi reconcile.Requests çıktısı veren bir dönüşüm fonksiyonu çalıştırarak İstekleri kuyruğa alır.
// reconcile.Requests, kaynak Olayın kullanıcı tarafından belirlenen bir dönüşümü ile tanımlanan
// rastgele bir nesne kümesi için olabilir. (örneğin, bir Node ekleyerek veya silerek tetiklenen bir küme yeniden boyutlandırma olayı
// için bir dizi nesne için Reconciler'ı tetikleyin)
//
// EnqueueRequestsFromMapFunc, bir nesneden bir veya daha fazla farklı türdeki nesneye güncellemeleri yaymak için sıkça kullanılır.
//
// Hem yeni hem de eski nesneyi içeren UpdateEvents için, dönüşüm fonksiyonu her iki nesne üzerinde de çalıştırılır ve her iki İstek kümesi de kuyruğa alınır.
func EnqueueRequestsFromMapFunc(fn MapFunc) EventHandler {
	return TypedEnqueueRequestsFromMapFunc(fn)
}

// TypedEnqueueRequestsFromMapFunc, her Olayda bir dizi reconcile.Requests çıktısı veren bir dönüşüm fonksiyonu çalıştırarak İstekleri kuyruğa alır.
// reconcile.Requests, kaynak Olayın kullanıcı tarafından belirlenen bir dönüşümü ile tanımlanan
// rastgele bir nesne kümesi için olabilir. (örneğin, bir Node ekleyerek veya silerek tetiklenen bir küme yeniden boyutlandırma olayı
// için bir dizi nesne için Reconciler'ı tetikleyin)
//
// TypedEnqueueRequestsFromMapFunc, bir nesneden bir veya daha fazla farklı türdeki nesneye güncellemeleri yaymak için sıkça kullanılır.
//
// Hem yeni hem de eski nesneyi içeren TypedUpdateEvents için, dönüşüm fonksiyonu her iki nesne üzerinde de çalıştırılır ve her iki İstek kümesi de kuyruğa alınır.
//
// TypedEnqueueRequestsFromMapFunc deneysel olup gelecekte değişikliğe tabidir.
func TypedEnqueueRequestsFromMapFunc[object any, request comparable](fn TypedMapFunc[object, request]) TypedEventHandler[object, request] {
	return &enqueueRequestsFromMapFunc[object, request]{
		toRequests: fn,
	}
}

var _ EventHandler = &enqueueRequestsFromMapFunc[client.Object, reconcile.Request]{}

type enqueueRequestsFromMapFunc[object any, request comparable] struct {
	// Mapper, argümanı reconcile edilecek anahtarlar dizisine dönüştürür
	toRequests TypedMapFunc[object, request]
}

// Create, EventHandler'ı uygular.
func (e *enqueueRequestsFromMapFunc[object, request]) Create(
	ctx context.Context,
	evt event.TypedCreateEvent[object],
	q workqueue.TypedRateLimitingInterface[request],
) {
	reqs := map[request]empty{}
	e.mapAndEnqueue(ctx, q, evt.Object, reqs)
}

// Update, EventHandler'ı uygular.
func (e *enqueueRequestsFromMapFunc[object, request]) Update(
	ctx context.Context,
	evt event.TypedUpdateEvent[object],
	q workqueue.TypedRateLimitingInterface[request],
) {
	reqs := map[request]empty{}
	e.mapAndEnqueue(ctx, q, evt.ObjectOld, reqs)
	e.mapAndEnqueue(ctx, q, evt.ObjectNew, reqs)
}

// Delete, EventHandler'ı uygular.
func (e *enqueueRequestsFromMapFunc[object, request]) Delete(
	ctx context.Context,
	evt event.TypedDeleteEvent[object],
	q workqueue.TypedRateLimitingInterface[request],
) {
	reqs := map[request]empty{}
	e.mapAndEnqueue(ctx, q, evt.Object, reqs)
}

// Generic, EventHandler'ı uygular.
func (e *enqueueRequestsFromMapFunc[object, request]) Generic(
	ctx context.Context,
	evt event.TypedGenericEvent[object],
	q workqueue.TypedRateLimitingInterface[request],
) {
	reqs := map[request]empty{}
	e.mapAndEnqueue(ctx, q, evt.Object, reqs)
}

func (e *enqueueRequestsFromMapFunc[object, request]) mapAndEnqueue(ctx context.Context, q workqueue.TypedRateLimitingInterface[request], o object, reqs map[request]empty) {
	for _, req := range e.toRequests(ctx, o) {
		_, ok := reqs[req]
		if !ok {
			q.Add(req)
			reqs[req] = empty{}
		}
	}
}
