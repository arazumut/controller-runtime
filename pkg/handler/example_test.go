/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği olmadıkça,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
haklar ve sınırlamalar için Lisansa bakınız.
*/

package handler_test

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	mgr manager.Manager
	c   controller.Controller
)

// Bu örnek, Pod'ları izler ve Olaydan (örneğin, Oluşturma, Güncelleme, Silme nedeniyle oluşan değişiklik) Pod'un Adı ve Namespace'ini içeren İstekleri sıraya alır.
func OrnekEnqueueRequestForObject() {
	// controller bir controller.controller
	err := c.Watch(
		source.Kind(mgr.GetCache(), &corev1.Pod{}, &handler.TypedEnqueueRequestForObject[*corev1.Pod]{}),
	)
	if err != nil {
		// hatayı ele al
	}
}

// Bu örnek, ReplicaSet'leri izler ve ReplicaSet'in oluşturulmasından sorumlu olan sahip (doğrudan) Deployment'ın Adı ve Namespace'ini içeren bir İstek sıraya alır.
func OrnekEnqueueRequestForOwner() {
	// controller bir controller.controller
	err := c.Watch(
		source.Kind(mgr.GetCache(),
			&appsv1.ReplicaSet{},
			handler.TypedEnqueueRequestForOwner[*appsv1.ReplicaSet](mgr.GetScheme(), mgr.GetRESTMapper(), &appsv1.Deployment{}, handler.OnlyControllerOwner()),
		),
	)
	if err != nil {
		// hatayı ele al
	}
}

// Bu örnek, Deployment'ları izler ve kullanıcı tarafından tanımlanan bir eşleme fonksiyonu kullanarak farklı nesnelerin (Tür: MyKind) Adı ve Namespace'ini içeren bir İstek sıraya alır.
func OrnekEnqueueRequestsFromMapFunc() {
	// controller bir controller.controller
	err := c.Watch(
		source.Kind(mgr.GetCache(), &appsv1.Deployment{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, a *appsv1.Deployment) []reconcile.Request {
				return []reconcile.Request{
					{NamespacedName: types.NamespacedName{
						Name:      a.Name + "-1",
						Namespace: a.Namespace,
					}},
					{NamespacedName: types.NamespacedName{
						Name:      a.Name + "-2",
						Namespace: a.Namespace,
					}},
				}
			}),
		),
	)
	if err != nil {
		// hatayı ele al
	}
}

// Bu örnek handler.EnqueueRequestForObject'u uygular.
func OrnekFuncs() {
	// controller bir controller.controller
	err := c.Watch(
		source.Kind(mgr.GetCache(), &corev1.Pod{},
			handler.TypedFuncs[*corev1.Pod, reconcile.Request]{
				CreateFunc: func(ctx context.Context, e event.TypedCreateEvent[*corev1.Pod], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
					q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
						Name:      e.Object.Name,
						Namespace: e.Object.Namespace,
					}})
				},
				UpdateFunc: func(ctx context.Context, e event.TypedUpdateEvent[*corev1.Pod], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
					q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
						Name:      e.ObjectNew.Name,
						Namespace: e.ObjectNew.Namespace,
					}})
				},
				DeleteFunc: func(ctx context.Context, e event.TypedDeleteEvent[*corev1.Pod], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
					q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
						Name:      e.Object.Name,
						Namespace: e.Object.Namespace,
					}})
				},
				GenericFunc: func(ctx context.Context, e event.TypedGenericEvent[*corev1.Pod], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
					q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
						Name:      e.Object.Name,
						Namespace: e.Object.Namespace,
					}})
				},
			},
		),
	)
	if err != nil {
		// hatayı ele al
	}
}
