/*
Kubernetes Yazarları 2019

Apache Lisansı, Sürüm 2.0 ("Lisans") kapsamında lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisans'a bakınız.
*/

package main

import (
	"context"
	"math/rand"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	api "sigs.k8s.io/controller-runtime/examples/crd/pkg"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var setupLog = ctrl.Log.WithName("setup")

// Reconciler yapısı, ChaosPod kaynaklarını yönetmek için kullanılır.
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile fonksiyonu, ChaosPod'ların yaşam döngüsünü yönetir.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("chaospod", req.NamespacedName)
	log.V(1).Info("ChaosPod Yeniden Uzlaştırılıyor")

	// ChaosPod kaynağını getir
	var chaospod api.ChaosPod
	if err := r.Get(ctx, req.NamespacedName, &chaospod); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("ChaosPod kaynağı bulunamadı. Silinmiş olmalı, göz ardı ediliyor.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "ChaosPod getirilemedi")
		return ctrl.Result{}, err
	}

	// İlgili Pod'un var olup olmadığını kontrol et
	var pod corev1.Pod
	podFound := true
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "İlgili Pod getirilemedi")
			return ctrl.Result{}, err
		}
		podFound = false
	}

	if podFound {
		// Pod'u durdurma zamanı gelip gelmediğini kontrol et
		shouldStop := chaospod.Spec.NextStop.Time.Before(time.Now())
		if !shouldStop {
			// NextStop zamanına kadar yeniden sıraya al
			return ctrl.Result{RequeueAfter: chaospod.Spec.NextStop.Sub(time.Now()) + 1*time.Second}, nil
		}

		// Pod'u durdurma zamanı geldiyse sil
		if err := r.Delete(ctx, &pod); err != nil {
			log.Error(err, "Pod silinemedi")
			return ctrl.Result{}, err
		}

		log.Info("Pod silindi, bir sonraki döngü için yeniden sıraya alınıyor")
		return ctrl.Result{Requeue: true}, nil
	}

	// Pod bulunamadıysa yeni bir Pod oluştur
	podTemplate := chaospod.Spec.Template.DeepCopy()
	pod.ObjectMeta = podTemplate.ObjectMeta
	pod.Name = req.Name
	pod.Namespace = req.Namespace
	pod.Spec = podTemplate.Spec

	if err := ctrl.SetControllerReference(&chaospod, &pod, r.Scheme); err != nil {
		log.Error(err, "Pod sahiplik referansı ayarlanamadı")
		return ctrl.Result{}, err
	}

	if err := r.Create(ctx, &pod); err != nil {
		log.Error(err, "Pod oluşturulamadı")
		return ctrl.Result{}, err
	}

	// NextStop zamanını ve ChaosPod durumunu güncelle
	chaospod.Spec.NextStop.Time = time.Now().Add(time.Duration(10*(rand.Int63n(2)+1)) * time.Second)
	chaospod.Status.LastRun = pod.CreationTimestamp
	if err := r.Update(ctx, &chaospod); err != nil {
		log.Error(err, "ChaosPod durumu güncellenemedi")
		return ctrl.Result{}, err
	}

	log.Info("Pod başarıyla oluşturuldu, bir sonraki durdurma için yeniden sıraya alınıyor")
	return ctrl.Result{}, nil
}

func main() {
	// Logger'ı ayarla
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	// Yeni bir yönetici oluştur
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		setupLog.Error(err, "Yönetici başlatılamıyor")
		os.Exit(1)
	}

	// Özel şemayı ekle (ChaosPod CRD)
	if err := api.AddToScheme(mgr.GetScheme()); err != nil {
		setupLog.Error(err, "ChaosPod şeması eklenemiyor")
		os.Exit(1)
	}

	// ChaosPod için yeni bir denetleyici oluştur
	err = ctrl.NewControllerManagedBy(mgr).
		For(&api.ChaosPod{}).
		Owns(&corev1.Pod{}).
		Complete(&Reconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		})
	if err != nil {
		setupLog.Error(err, "Denetleyici oluşturulamıyor")
		os.Exit(1)
	}

	// Yönetici başlat
	setupLog.Info("Yönetici başlatılıyor")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "Yönetici çalıştırılırken sorun oluştu")
		os.Exit(1)
	}
}
