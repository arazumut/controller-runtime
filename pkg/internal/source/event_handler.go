/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisansa bakınız.
*/

package internal

import (
	"context"
	"fmt"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/internal/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var log = logf.RuntimeLog.WithName("source").WithName("EventHandler")

// NewEventHandler yeni bir EventHandler oluşturur.
func NewEventHandler[object client.Object, request comparable](
	ctx context.Context,
	queue workqueue.TypedRateLimitingInterface[request],
	handler handler.TypedEventHandler[object, request],
	predicates []predicate.TypedPredicate[object]) *EventHandler[object, request] {
	return &EventHandler[object, request]{
		ctx:        ctx,
		handler:    handler,
		queue:      queue,
		predicates: predicates,
	}
}

// EventHandler, handler.EventHandler arayüzünü cache.ResourceEventHandler arayüzüne adapte eder.
type EventHandler[object client.Object, request comparable] struct {
	// ctx, olay işleyicisini oluşturan bağlamı saklar
	// iptal sinyallerini her işleyici işlevine yaymak için kullanılır.
	ctx context.Context

	handler    handler.TypedEventHandler[object, request]
	queue      workqueue.TypedRateLimitingInterface[request]
	predicates []predicate.TypedPredicate[object]
}

// HandlerFuncs, EventHandler'ı ResourceEventHandlerFuncs'e dönüştürür
// TODO: client-go 1.27 ile ResourceEventHandlerDetailedFuncs'e geçiş yap
func (e *EventHandler[object, request]) HandlerFuncs() cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    e.OnAdd,
		UpdateFunc: e.OnUpdate,
		DeleteFunc: e.OnDelete,
	}
}

// OnAdd, CreateEvent oluşturur ve EventHandler'da Create'i çağırır.
func (e *EventHandler[object, request]) OnAdd(obj interface{}) {
	c := event.TypedCreateEvent[object]{}

	// Nesneyi objeden çıkar
	if o, ok := obj.(object); ok {
		c.Object = o
	} else {
		log.Error(nil, "OnAdd eksik Nesne",
			"nesne", obj, "tip", fmt.Sprintf("%T", obj))
		return
	}

	for _, p := range e.predicates {
		if !p.Create(c) {
			return
		}
	}

	// Oluşturma işleyicisini çağır
	ctx, cancel := context.WithCancel(e.ctx)
	defer cancel()
	e.handler.Create(ctx, c, e.queue)
}

// OnUpdate, UpdateEvent oluşturur ve EventHandler'da Update'i çağırır.
func (e *EventHandler[object, request]) OnUpdate(oldObj, newObj interface{}) {
	u := event.TypedUpdateEvent[object]{}

	if o, ok := oldObj.(object); ok {
		u.ObjectOld = o
	} else {
		log.Error(nil, "OnUpdate eksik ObjectOld",
			"nesne", oldObj, "tip", fmt.Sprintf("%T", oldObj))
		return
	}

	// Nesneyi objeden çıkar
	if o, ok := newObj.(object); ok {
		u.ObjectNew = o
	} else {
		log.Error(nil, "OnUpdate eksik ObjectNew",
			"nesne", newObj, "tip", fmt.Sprintf("%T", newObj))
		return
	}

	for _, p := range e.predicates {
		if !p.Update(u) {
			return
		}
	}

	// Güncelleme işleyicisini çağır
	ctx, cancel := context.WithCancel(e.ctx)
	defer cancel()
	e.handler.Update(ctx, u, e.queue)
}

// OnDelete, DeleteEvent oluşturur ve EventHandler'da Delete'i çağırır.
func (e *EventHandler[object, request]) OnDelete(obj interface{}) {
	d := event.TypedDeleteEvent[object]{}

	// Tombstone olaylarıyla başa çıkmak için nesneyi çıkar.
	// Tombstone olayları nesneyi DeleteFinalStateUnknown yapısında sarar, bu yüzden nesne çıkarılmalıdır.
	// sample-controller'dan kopyalandı
	// Bu, olayları kaçırmadığımız sürece asla olmamalıdır, ki olmadığımızı ve bu inanç üzerine kararlar aldığımızı kabul ettik. Belki bu burada olmamalı?
	var ok bool
	if _, ok = obj.(client.Object); !ok {
		// Nesne Metadata'ya sahip değilse, bunun DeletedFinalStateUnknown türünde bir tombstone nesnesi olduğunu varsay.
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			log.Error(nil, "Nesneleri çözümlemede hata. cache.DeletedFinalStateUnknown bekleniyordu",
				"tip", fmt.Sprintf("%T", obj),
				"nesne", obj)
			return
		}

		// DeleteStateUnknown'u true olarak ayarla
		d.DeleteStateUnknown = true

		// obj'yi tombstone obj'ye ayarla
		obj = tombstone.Obj
	}

	// Nesneyi objeden çıkar
	if o, ok := obj.(object); ok {
		d.Object = o
	} else {
		log.Error(nil, "OnDelete eksik Nesne",
			"nesne", obj, "tip", fmt.Sprintf("%T", obj))
		return
	}

	for _, p := range e.predicates {
		if !p.Delete(d) {
			return
		}
	}

	// Silme işleyicisini çağır
	ctx, cancel := context.WithCancel(e.ctx)
	defer cancel()
	e.handler.Delete(ctx, d, e.queue)
}
