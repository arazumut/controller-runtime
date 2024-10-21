/*
2018 Kubernetes Yazarları tarafından oluşturulmuştur.

Apache Lisansı, Sürüm 2.0 ("Lisans") kapsamında lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izinle aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
haklar ve sınırlamalar için Lisansa bakınız.
*/

package controller_test

import (
	"context"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	mgr manager.Manager
	// Not: init() içinde SetLogger çağırmayın, aksi takdirde ana suite'deki loglamayı bozarsınız.
	log = logf.Log.WithName("controller-examples")
)

// Bu örnek, "pod-controller" adında yeni bir Controller oluşturur ve no-op reconcile fonksiyonu ile başlatır.
// manager.Manager, Controller'ı başlatmak için kullanılacak ve ona paylaşılan bir Cache ve Client sağlayacaktır.
func OrnekYeni() {
	_, err := controller.New("pod-controller", mgr, controller.Options{
		Reconciler: reconcile.Func(func(context.Context, reconcile.Request) (reconcile.Result, error) {
			// API'yi oluşturma, güncelleme, silme işlemleri ile uygulamak için iş mantığınız buraya gelir.
			return reconcile.Result{}, nil
		}),
	})
	if err != nil {
		log.Error(err, "pod-controller oluşturulamadı")
		os.Exit(1)
	}
}

// Bu örnek, Pod'ları İzlemek ve no-op Reconciler'ı çağırmak için "pod-controller" adında yeni bir Controller başlatır.
func OrnekController() {
	// mgr bir manager.Manager'dır

	// Sağlanan Reconciler fonksiyonunu olaylara yanıt olarak çağıracak yeni bir Controller oluşturun.
	c, err := controller.New("pod-controller", mgr, controller.Options{
		Reconciler: reconcile.Func(func(context.Context, reconcile.Request) (reconcile.Result, error) {
			// API'yi oluşturma, güncelleme, silme işlemleri ile uygulamak için iş mantığınız buraya gelir.
			return reconcile.Result{}, nil
		}),
	})
	if err != nil {
		log.Error(err, "pod-controller oluşturulamadı")
		os.Exit(1)
	}

	// Pod oluşturma / güncelleme / silme olaylarını izleyin ve Reconcile çağırın
	err = c.Watch(source.Kind(mgr.GetCache(), &corev1.Pod{}, &handler.TypedEnqueueRequestForObject[*corev1.Pod]{}))
	if err != nil {
		log.Error(err, "pod'ları izleyemedi")
		os.Exit(1)
	}

	// Controller'ı manager üzerinden başlatın.
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "manager çalışmaya devam edemedi")
		os.Exit(1)
	}
}

// Bu örnek, yapılandırılmamış nesne ile Pod'ları İzlemek ve no-op Reconciler'ı çağırmak için "pod-controller" adında yeni bir Controller başlatır.
func OrnekController_yapilandirilmamis() {
	// mgr bir manager.Manager'dır

	// Sağlanan Reconciler fonksiyonunu olaylara yanıt olarak çağıracak yeni bir Controller oluşturun.
	c, err := controller.New("pod-controller", mgr, controller.Options{
		Reconciler: reconcile.Func(func(context.Context, reconcile.Request) (reconcile.Result, error) {
			// API'yi oluşturma, güncelleme, silme işlemleri ile uygulamak için iş mantığınız buraya gelir.
			return reconcile.Result{}, nil
		}),
	})
	if err != nil {
		log.Error(err, "pod-controller oluşturulamadı")
		os.Exit(1)
	}

	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Kind:    "Pod",
		Group:   "",
		Version: "v1",
	})
	// Pod oluşturma / güncelleme / silme olaylarını izleyin ve Reconcile çağırın
	err = c.Watch(source.Kind(mgr.GetCache(), u, &handler.TypedEnqueueRequestForObject[*unstructured.Unstructured]{}))
	if err != nil {
		log.Error(err, "pod'ları izleyemedi")
		os.Exit(1)
	}

	// Controller'ı manager üzerinden başlatın.
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "manager çalışmaya devam edemedi")
		os.Exit(1)
	}
}

// Bu örnek, Pod'ları izlemek ve no-op reconciler'ı çağırmak için "pod-controller" adında yeni bir controller oluşturur.
// Controller sağlanan manager'a eklenmez ve bu nedenle çağıran tarafından başlatılmalı ve durdurulmalıdır.
func OrnekYeniYonetimsiz() {
	// mgr bir manager.Manager'dır

	// Sağlanan manager'a eklenmeyen yeni bir controller oluşturur.
	c, err := controller.NewUnmanaged("pod-controller", mgr, controller.Options{
		Reconciler: reconcile.Func(func(context.Context, reconcile.Request) (reconcile.Result, error) {
			return reconcile.Result{}, nil
		}),
	})
	if err != nil {
		log.Error(err, "pod-controller oluşturulamadı")
		os.Exit(1)
	}

	if err := c.Watch(source.Kind(mgr.GetCache(), &corev1.Pod{}, &handler.TypedEnqueueRequestForObject[*corev1.Pod]{})); err != nil {
		log.Error(err, "pod'ları izleyemedi")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Controller'ımızı bir goroutine içinde başlatın, böylece engellenmeyiz.
	go func() {
		// Controller manager'ımız lider olarak seçilene kadar bekleyin. Tüm sürecimizin liderliği kaybedersek sona ereceğini varsayıyoruz, bu yüzden bunu ele almamız gerekmiyor.
		<-mgr.Elected()

		// Controller'ımızı başlatın. Bu, context kapatılana veya controller bir hata döndürene kadar engellenecektir.
		if err := c.Start(ctx); err != nil {
			log.Error(err, "deney controller'ı çalıştırılamıyor")
		}
	}()

	// Controller'ımızı durdurun.
	cancel()
}
