/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa uyarınca veya yazılı olarak kabul edilmediği sürece,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisansa bakın.
*/

package handler

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/internal/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ EventHandler = &enqueueRequestForOwner[client.Object]{}

var log = logf.RuntimeLog.WithName("eventhandler").WithName("enqueueRequestForOwner")

// OwnerOption, EnqueueRequestForOwner EventHandler'ını değiştirir.
type OwnerOption func(e enqueueRequestForOwnerInterface)

// EnqueueRequestForOwner, bir nesnenin sahipleri için İstekleri sıraya alır. Örneğin, olayın kaynağı olan nesneyi oluşturan nesne.
//
// Eğer bir ReplicaSet Pod'lar oluşturursa, kullanıcılar Pod Olaylarına yanıt olarak ReplicaSet'i reconcile edebilirler:
//
// - Pod türünde bir source.Kind Kaynağı.
//
// - ReplicaSet türünde bir OwnerType ve OnlyControllerOwner ayarı true olan bir handler.enqueueRequestForOwner EventHandler.
func EnqueueRequestForOwner(scheme *runtime.Scheme, mapper meta.RESTMapper, ownerType client.Object, opts ...OwnerOption) EventHandler {
	return TypedEnqueueRequestForOwner[client.Object](scheme, mapper, ownerType, opts...)
}

// TypedEnqueueRequestForOwner, bir nesnenin sahipleri için İstekleri sıraya alır. Örneğin, olayın kaynağı olan nesneyi oluşturan nesne.
//
// Eğer bir ReplicaSet Pod'lar oluşturursa, kullanıcılar Pod Olaylarına yanıt olarak ReplicaSet'i reconcile edebilirler:
//
// - Pod türünde bir source.Kind Kaynağı.
//
// - ReplicaSet türünde bir OwnerType ve OnlyControllerOwner ayarı true olan bir handler.typedEnqueueRequestForOwner EventHandler.
//
// TypedEnqueueRequestForOwner deneysel olup gelecekte değişikliğe tabidir.
func TypedEnqueueRequestForOwner[object client.Object](scheme *runtime.Scheme, mapper meta.RESTMapper, ownerType client.Object, opts ...OwnerOption) TypedEventHandler[object, reconcile.Request] {
	e := &enqueueRequestForOwner[object]{
		ownerType: ownerType,
		mapper:    mapper,
	}
	if err := e.parseOwnerTypeGroupKind(scheme); err != nil {
		panic(err)
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// OnlyControllerOwner sağlanırsa, yalnızca Controller: true olan ilk OwnerReference'a bakar.
func OnlyControllerOwner() OwnerOption {
	return func(e enqueueRequestForOwnerInterface) {
		e.setIsController(true)
	}
}

type enqueueRequestForOwnerInterface interface {
	setIsController(bool)
}

type enqueueRequestForOwner[object client.Object] struct {
	// ownerType, OwnerReferences'de aramak için Sahip nesnesinin türüdür. Yalnızca Grup ve Tür karşılaştırılır.
	ownerType runtime.Object

	// isController ayarlanmışsa, yalnızca Controller: true olan ilk OwnerReference'a bakar.
	isController bool

	// groupKind, OwnerType'dan önbelleğe alınan Grup ve Tür'dür.
	groupKind schema.GroupKind

	// mapper, GroupVersionKinds'ı Kaynaklara eşler.
	mapper meta.RESTMapper
}

func (e *enqueueRequestForOwner[object]) setIsController(isController bool) {
	e.isController = isController
}

// Create, EventHandler'ı uygular.
func (e *enqueueRequestForOwner[object]) Create(ctx context.Context, evt event.TypedCreateEvent[object], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	reqs := map[reconcile.Request]empty{}
	e.getOwnerReconcileRequest(evt.Object, reqs)
	for req := range reqs {
		q.Add(req)
	}
}

// Update, EventHandler'ı uygular.
func (e *enqueueRequestForOwner[object]) Update(ctx context.Context, evt event.TypedUpdateEvent[object], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	reqs := map[reconcile.Request]empty{}
	e.getOwnerReconcileRequest(evt.ObjectOld, reqs)
	e.getOwnerReconcileRequest(evt.ObjectNew, reqs)
	for req := range reqs {
		q.Add(req)
	}
}

// Delete, EventHandler'ı uygular.
func (e *enqueueRequestForOwner[object]) Delete(ctx context.Context, evt event.TypedDeleteEvent[object], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	reqs := map[reconcile.Request]empty{}
	e.getOwnerReconcileRequest(evt.Object, reqs)
	for req := range reqs {
		q.Add(req)
	}
}

// Generic, EventHandler'ı uygular.
func (e *enqueueRequestForOwner[object]) Generic(ctx context.Context, evt event.TypedGenericEvent[object], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	reqs := map[reconcile.Request]empty{}
	e.getOwnerReconcileRequest(evt.Object, reqs)
	for req := range reqs {
		q.Add(req)
	}
}

// parseOwnerTypeGroupKind, OwnerType'ı bir Grup ve Tür olarak ayrıştırır ve sonucu önbelleğe alır.
// OwnerType, şema kullanılarak ayrıştırılamazsa false döner.
func (e *enqueueRequestForOwner[object]) parseOwnerTypeGroupKind(scheme *runtime.Scheme) error {
	// Türlerin türlerini alın
	kinds, _, err := scheme.ObjectKinds(e.ownerType)
	if err != nil {
		log.Error(err, "OwnerType için ObjectKinds alınamadı", "owner type", fmt.Sprintf("%T", e.ownerType))
		return err
	}
	// Yalnızca 1 tür bekleyin. Birden fazla tür varsa, bu muhtemelen ListOptions gibi bir kenar durumudur.
	if len(kinds) != 1 {
		err := fmt.Errorf("OwnerType %T için tam olarak 1 tür bekleniyordu, ancak %s tür bulundu", e.ownerType, kinds)
		log.Error(nil, "OwnerType için tam olarak 1 tür bekleniyordu", "owner type", fmt.Sprintf("%T", e.ownerType), "kinds", kinds)
		return err
	}
	// OwnerType için Grup ve Tür'ü önbelleğe alın
	e.groupKind = schema.GroupKind{Group: kinds[0].Group, Kind: kinds[0].Kind}
	return nil
}

// getOwnerReconcileRequest, nesneye bakar ve e.OwnerType ile eşleşen nesnenin sahiplerine reconcile.Request haritası oluşturur.
func (e *enqueueRequestForOwner[object]) getOwnerReconcileRequest(obj metav1.Object, result map[reconcile.Request]empty) {
	// Kullanıcı tarafından istenen OwnerType'ın Grup ve Tür'ü ile eşleşen OwnerReferences'ı arayarak yineleyin
	for _, ref := range e.getOwnersReferences(obj) {
		// OwnerReference'dan Grubu ayrıştırarak kullanıcı tarafından istenen OwnerType'dan ayrıştırılan Grup ile karşılaştırın
		refGV, err := schema.ParseGroupVersion(ref.APIVersion)
		if err != nil {
			log.Error(err, "OwnerReference APIVersion ayrıştırılamadı",
				"api version", ref.APIVersion)
			return
		}

		// OwnerReference Grup ve Tür'ü, kullanıcı tarafından belirtilen OwnerType Grup ve Tür ile karşılaştırın.
		// İki eşleşirse, OwnerReference'da belirtilen nesne için bir İstek oluşturun.
		// OwnerReference'dan Adı ve olaydaki nesneden Namespace'i kullanın.
		if ref.Kind == e.groupKind.Kind && refGV.Group == e.groupKind.Group {
			// Eşleşme bulundu - OwnerReference'da belirtilen nesne için bir İstek ekleyin
			request := reconcile.Request{NamespacedName: types.NamespacedName{
				Name: ref.Name,
			}}

			// Eğer sahip namespaced değilse, namespace'i ayarlamamalıyız
			mapping, err := e.mapper.RESTMapping(e.groupKind, refGV.Version)
			if err != nil {
				log.Error(err, "rest mapping alınamadı", "kind", e.groupKind)
				return
			}
			if mapping.Scope.Name() != meta.RESTScopeNameRoot {
				request.Namespace = obj.GetNamespace()
			}

			result[request] = empty{}
		}
	}
}

// getOwnersReferences, enqueueRequestForOwner tarafından belirtilen bir nesne için OwnerReferences'ı döndürür
// - IsController true ise: yalnızca Controller OwnerReference'ı alın (bulunursa)
// - IsController false ise: tüm OwnerReferences'ı alın.
func (e *enqueueRequestForOwner[object]) getOwnersReferences(obj metav1.Object) []metav1.OwnerReference {
	if obj == nil {
		return nil
	}

	// Controller olarak filtrelenmemişse, tüm OwnerReferences'ı kullanın
	if !e.isController {
		return obj.GetOwnerReferences()
	}
	// Controller olarak filtrelenmişse, yalnızca Controller OwnerReference'ı alın
	if ownerRef := metav1.GetControllerOf(obj); ownerRef != nil {
		return []metav1.OwnerReference{*ownerRef}
	}
	// Controller OwnerReference bulunamadı
	return nil
}
