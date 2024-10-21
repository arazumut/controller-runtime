/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisansa bakınız.
*/

package builder_test

import (
	"context"
	"fmt"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func OrnekBuilder_metadata_only() {
	logf.SetLogger(zap.New())

	log := logf.Log.WithName("builder-ornekleri")

	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		log.Error(err, "yönetici oluşturulamadı")
		os.Exit(1)
	}

	cl := mgr.GetClient()
	err = builder.
		ControllerManagedBy(mgr).                  // ControllerManagedBy oluştur
		For(&appsv1.ReplicaSet{}).                 // ReplicaSet, Uygulama API'sidir
		Owns(&corev1.Pod{}, builder.OnlyMetadata). // ReplicaSet, oluşturduğu Pod'lara sahiptir ve bunları yalnızca meta veri olarak önbelleğe alır
		Complete(reconcile.Func(func(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
			// ReplicaSet'i oku
			rs := &appsv1.ReplicaSet{}
			err := cl.Get(ctx, req.NamespacedName, rs)
			if err != nil {
				return reconcile.Result{}, client.IgnoreNotFound(err)
			}

			// PodTemplate Etiketleri ile eşleşen Pod'ların yalnızca meta verilerini listele
			var podsMeta metav1.PartialObjectMetadataList
			err = cl.List(ctx, &podsMeta, client.InNamespace(req.Namespace), client.MatchingLabels(rs.Spec.Template.Labels))
			if err != nil {
				return reconcile.Result{}, client.IgnoreNotFound(err)
			}

			// ReplicaSet'i güncelle
			rs.Labels["pod-count"] = fmt.Sprintf("%v", len(podsMeta.Items))
			err = cl.Update(ctx, rs)
			if err != nil {
				return reconcile.Result{}, err
			}

			return reconcile.Result{}, nil
		}))
	if err != nil {
		log.Error(err, "kontrolcü oluşturulamadı")
		os.Exit(1)
	}

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "yönetici başlatılamadı")
		os.Exit(1)
	}
}

// Bu örnek, ReplicaSets ve Pod'lar için yapılandırılmış basit bir uygulama ControllerManagedBy oluşturur.
//
// * ReplicaSets için yeni bir uygulama oluşturun ve ReplicaSetReconciler'a çağrı yaparak
// ReplicaSet tarafından oluşturulan Pod'ları yönetin.
//
// * Uygulamayı başlatın.
func OrnekBuilder() {
	logf.SetLogger(zap.New())

	log := logf.Log.WithName("builder-ornekleri")

	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		log.Error(err, "yönetici oluşturulamadı")
		os.Exit(1)
	}

	err = builder.
		ControllerManagedBy(mgr).  // ControllerManagedBy oluştur
		For(&appsv1.ReplicaSet{}). // ReplicaSet, Uygulama API'sidir
		Owns(&corev1.Pod{}).       // ReplicaSet, oluşturduğu Pod'lara sahiptir
		Complete(&ReplicaSetReconciler{
			Client: mgr.GetClient(),
		})
	if err != nil {
		log.Error(err, "kontrolcü oluşturulamadı")
		os.Exit(1)
	}

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "yönetici başlatılamadı")
		os.Exit(1)
	}
}

// ReplicaSetReconciler, basit bir ControllerManagedBy örnek uygulamasıdır.
type ReplicaSetReconciler struct {
	client.Client
}

// İş mantığını uygula:
// Bu işlev, bir ReplicaSet veya bir ReplicaSet'e sahip bir Pod'da değişiklik olduğunda çağrılacaktır.
//
// * ReplicaSet'i oku
// * Pod'ları oku
// * Pod sayısı ile ReplicaSet üzerinde bir Etiket ayarla.
func (a *ReplicaSetReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	// ReplicaSet'i oku
	rs := &appsv1.ReplicaSet{}
	err := a.Get(ctx, req.NamespacedName, rs)
	if err != nil {
		return reconcile.Result{}, err
	}

	// PodTemplate Etiketleri ile eşleşen Pod'ları listele
	pods := &corev1.PodList{}
	err = a.List(ctx, pods, client.InNamespace(req.Namespace), client.MatchingLabels(rs.Spec.Template.Labels))
	if err != nil {
		return reconcile.Result{}, err
	}

	// ReplicaSet'i güncelle
	rs.Labels["pod-count"] = fmt.Sprintf("%v", len(pods.Items))
	err = a.Update(ctx, rs)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
