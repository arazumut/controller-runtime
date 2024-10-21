/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamındaki izinleri ve sınırlamaları yöneten özel dil için
Lisans'a bakınız.
*/

package handler

import (
	"context"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// EventHandler, olaylara (örneğin, Pod Oluşturma) yanıt olarak reconcile.Request'leri sıraya alır. EventHandler'lar bir nesne için bir Olayı
// aynı nesne veya farklı nesneler için reconcile tetiklemek üzere eşler - örneğin, Foo türünde bir nesne için bir Olay varsa (source.Kind kullanarak),
// Bar türünde bir veya daha fazla nesneyi reconcile edin.
//
// Aynı reconcile.Request'ler, reconcile çağrılmadan önce sıralama mekanizması aracılığıyla bir araya getirilir.
//
// * Olayın olduğu nesneyi reconcile etmek için EnqueueRequestForObject kullanın
// - bu, Controller'ın reconcile ettiği türler için olaylar için yapılır. (örneğin, Deployment Controller için Deployment)
//
// * Olayın olduğu nesnenin sahibini reconcile etmek için EnqueueRequestForOwner kullanın
// - bu, Controller'ın oluşturduğu türler için olaylar için yapılır. (örneğin, Deployment Controller tarafından oluşturulan ReplicaSets)
//
// * Bir nesne için bir olayı farklı bir türde bir nesnenin reconcile'ına dönüştürmek için EnqueueRequestsFromMapFunc kullanın
// - bu, Controller'ın ilgilenebileceği ancak oluşturmadığı türler için olaylar için yapılır.
// (örneğin, Foo, küme boyutu olaylarına yanıt veriyorsa, Node olaylarını Foo nesnelerine eşleyin.)
//
// Kendi EventHandler'ınızı uygulamıyorsanız, EventHandler arayüzündeki işlevleri görmezden gelebilirsiniz.
// Çoğu kullanıcı kendi EventHandler'ını uygulamak zorunda kalmamalıdır.
type EventHandler = TypedEventHandler[client.Object, reconcile.Request]

// TypedEventHandler, olaylara (örneğin, Pod Oluşturma) yanıt olarak reconcile.Request'leri sıraya alır. TypedEventHandler'lar bir nesne için bir Olayı
// aynı nesne veya farklı nesneler için reconcile tetiklemek üzere eşler - örneğin, Foo türünde bir nesne için bir Olay varsa (source.Kind kullanarak),
// Bar türünde bir veya daha fazla nesneyi reconcile edin.
//
// Aynı reconcile.Request'ler, reconcile çağrılmadan önce sıralama mekanizması aracılığıyla bir araya getirilir.
//
// * Olayın olduğu nesneyi reconcile etmek için TypedEnqueueRequestForObject kullanın
// - bu, Controller'ın reconcile ettiği türler için olaylar için yapılır. (örneğin, Deployment Controller için Deployment)
//
// * Olayın olduğu nesnenin sahibini reconcile etmek için TypedEnqueueRequestForOwner kullanın
// - bu, Controller'ın oluşturduğu türler için olaylar için yapılır. (örneğin, Deployment Controller tarafından oluşturulan ReplicaSets)
//
// * Bir nesne için bir olayı farklı bir türde bir nesnenin reconcile'ına dönüştürmek için TypedEnqueueRequestsFromMapFunc kullanın
// - bu, Controller'ın ilgilenebileceği ancak oluşturmadığı türler için olaylar için yapılır.
// (örneğin, Foo, küme boyutu olaylarına yanıt veriyorsa, Node olaylarını Foo nesnelerine eşleyin.)
//
// Kendi TypedEventHandler'ınızı uygulamıyorsanız, TypedEventHandler arayüzündeki işlevleri görmezden gelebilirsiniz.
// Çoğu kullanıcı kendi TypedEventHandler'ını uygulamak zorunda kalmamalıdır.
//
// TypedEventHandler deneysel olup gelecekte değişikliğe tabidir.
type TypedEventHandler[object any, request comparable] interface {
	// Create, bir oluşturma olayına yanıt olarak çağrılır - örneğin, Pod Oluşturma.
	Create(context.Context, event.TypedCreateEvent[object], workqueue.TypedRateLimitingInterface[request])

	// Update, bir güncelleme olayına yanıt olarak çağrılır - örneğin, Pod Güncelleme.
	Update(context.Context, event.TypedUpdateEvent[object], workqueue.TypedRateLimitingInterface[request])

	// Delete, bir silme olayına yanıt olarak çağrılır - örneğin, Pod Silme.
	Delete(context.Context, event.TypedDeleteEvent[object], workqueue.TypedRateLimitingInterface[request])

	// Generic, bilinmeyen türde bir olaya veya bir cron veya dış tetikleyici istek olarak tetiklenen sentetik bir olaya yanıt olarak çağrılır
	// - örneğin, Autoscaling reconcile veya bir Webhook.
	Generic(context.Context, event.TypedGenericEvent[object], workqueue.TypedRateLimitingInterface[request])
}

var _ EventHandler = Funcs{}

// Funcs, eventhandler'ı uygular.
type Funcs = TypedFuncs[client.Object, reconcile.Request]

// TypedFuncs, eventhandler'ı uygular.
//
// TypedFuncs deneysel olup gelecekte değişikliğe tabidir.
type TypedFuncs[object any, request comparable] struct {
	// Create, bir ekleme olayına yanıt olarak çağrılır. Varsayılan olarak no-op.
	// RateLimitingInterface, reconcile.Request'leri sıraya almak için kullanılır.
	CreateFunc func(context.Context, event.TypedCreateEvent[object], workqueue.TypedRateLimitingInterface[request])

	// Update, bir güncelleme olayına yanıt olarak çağrılır. Varsayılan olarak no-op.
	// RateLimitingInterface, reconcile.Request'leri sıraya almak için kullanılır.
	UpdateFunc func(context.Context, event.TypedUpdateEvent[object], workqueue.TypedRateLimitingInterface[request])

	// Delete, bir silme olayına yanıt olarak çağrılır. Varsayılan olarak no-op.
	// RateLimitingInterface, reconcile.Request'leri sıraya almak için kullanılır.
	DeleteFunc func(context.Context, event.TypedDeleteEvent[object], workqueue.TypedRateLimitingInterface[request])

	// GenericFunc, genel bir olaya yanıt olarak çağrılır. Varsayılan olarak no-op.
	// RateLimitingInterface, reconcile.Request'leri sıraya almak için kullanılır.
	GenericFunc func(context.Context, event.TypedGenericEvent[object], workqueue.TypedRateLimitingInterface[request])
}

// Create, EventHandler'ı uygular.
func (h TypedFuncs[object, request]) Create(ctx context.Context, e event.TypedCreateEvent[object], q workqueue.TypedRateLimitingInterface[request]) {
	if h.CreateFunc != nil {
		h.CreateFunc(ctx, e, q)
	}
}

// Delete, EventHandler'ı uygular.
func (h TypedFuncs[object, request]) Delete(ctx context.Context, e event.TypedDeleteEvent[object], q workqueue.TypedRateLimitingInterface[request]) {
	if h.DeleteFunc != nil {
		h.DeleteFunc(ctx, e, q)
	}
}

// Update, EventHandler'ı uygular.
func (h TypedFuncs[object, request]) Update(ctx context.Context, e event.TypedUpdateEvent[object], q workqueue.TypedRateLimitingInterface[request]) {
	if h.UpdateFunc != nil {
		h.UpdateFunc(ctx, e, q)
	}
}

// Generic, EventHandler'ı uygular.
func (h TypedFuncs[object, request]) Generic(ctx context.Context, e event.TypedGenericEvent[object], q workqueue.TypedRateLimitingInterface[request]) {
	if h.GenericFunc != nil {
		h.GenericFunc(ctx, e, q)
	}
}
