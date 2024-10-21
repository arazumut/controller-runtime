/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisansa uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
yetkiler ve sınırlamalar için Lisansa bakınız.
*/

package handler

import (
	"context"
	"reflect"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/internal/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var enqueueLog = logf.RuntimeLog.WithName("eventhandler").WithName("EnqueueRequestForObject")

type empty struct{}

// EnqueueRequestForObject, olayın kaynağı olan nesnenin Adı ve Namespace'ini içeren bir İstek kuyruğa alır.
// (örneğin, oluşturulan / silinen / güncellenen nesnelerin Adı ve Namespace'i). handler.EnqueueRequestForObject, ilişkili
// Kaynakları (örneğin CRD'ler) olan hemen hemen tüm Kontrolörler tarafından ilişkili Kaynağı yeniden uzlaştırmak için kullanılır.
type EnqueueRequestForObject = TypedEnqueueRequestForObject[client.Object]

// TypedEnqueueRequestForObject, olayın kaynağı olan nesnenin Adı ve Namespace'ini içeren bir İstek kuyruğa alır.
// (örneğin, oluşturulan / silinen / güncellenen nesnelerin Adı ve Namespace'i). handler.TypedEnqueueRequestForObject, ilişkili
// Kaynakları (örneğin CRD'ler) olan hemen hemen tüm Kontrolörler tarafından ilişkili Kaynağı yeniden uzlaştırmak için kullanılır.
//
// TypedEnqueueRequestForObject deneysel olup gelecekte değişikliğe tabidir.
type TypedEnqueueRequestForObject[object client.Object] struct{}

// Create, EventHandler'ı uygular.
func (e *TypedEnqueueRequestForObject[T]) Create(ctx context.Context, evt event.TypedCreateEvent[T], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	if isNil(evt.Object) {
		enqueueLog.Error(nil, "CreateEvent, metadata olmadan alındı", "event", evt)
		return
	}
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      evt.Object.GetName(),
		Namespace: evt.Object.GetNamespace(),
	}})
}

// Update, EventHandler'ı uygular.
func (e *TypedEnqueueRequestForObject[T]) Update(ctx context.Context, evt event.TypedUpdateEvent[T], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	switch {
	case !isNil(evt.ObjectNew):
		q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
			Name:      evt.ObjectNew.GetName(),
			Namespace: evt.ObjectNew.GetNamespace(),
		}})
	case !isNil(evt.ObjectOld):
		q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
			Name:      evt.ObjectOld.GetName(),
			Namespace: evt.ObjectOld.GetNamespace(),
		}})
	default:
		enqueueLog.Error(nil, "UpdateEvent, metadata olmadan alındı", "event", evt)
	}
}

// Delete, EventHandler'ı uygular.
func (e *TypedEnqueueRequestForObject[T]) Delete(ctx context.Context, evt event.TypedDeleteEvent[T], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	if isNil(evt.Object) {
		enqueueLog.Error(nil, "DeleteEvent, metadata olmadan alındı", "event", evt)
		return
	}
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      evt.Object.GetName(),
		Namespace: evt.Object.GetNamespace(),
	}})
}

// Generic, EventHandler'ı uygular.
func (e *TypedEnqueueRequestForObject[T]) Generic(ctx context.Context, evt event.TypedGenericEvent[T], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	if isNil(evt.Object) {
		enqueueLog.Error(nil, "GenericEvent, metadata olmadan alındı", "event", evt)
		return
	}
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      evt.Object.GetName(),
		Namespace: evt.Object.GetNamespace(),
	}})
}

func isNil(arg any) bool {
	v := reflect.ValueOf(arg)
	return !v.IsValid() || ((v.Kind() == reflect.Ptr ||
		v.Kind() == reflect.Interface ||
		v.Kind() == reflect.Slice ||
		v.Kind() == reflect.Map ||
		v.Kind() == reflect.Chan ||
		v.Kind() == reflect.Func) && v.IsNil())
}
