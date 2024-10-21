/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamındaki izinler ve sınırlamalar hakkında daha fazla bilgi için
Lisans'a bakınız.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	"sigs.k8s.io/controller-runtime/pkg/internal/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Options yeni bir Controller oluşturmak için kullanılan argümanlardır.
type Options = TypedOptions[reconcile.Request]

// TypedOptions yeni bir Controller oluşturmak için kullanılan argümanlardır.
type TypedOptions[request comparable] struct {
	// SkipNameValidation, her denetleyici adının benzersiz olmasını sağlayan ad doğrulamasını atlamaya izin verir.
	// Benzersiz denetleyici adları, bir denetleyici için benzersiz metrikler ve günlükler almak için önemlidir.
	// Ayarlanmamışsa, Yöneticiden Controller.SkipNameValidation ayarına varsayılan olarak ayarlanır.
	// Yöneticiden Controller.SkipNameValidation ayarı da ayarlanmamışsa varsayılan olarak false olur.
	SkipNameValidation *bool

	// MaxConcurrentReconciles, çalıştırılabilecek maksimum eşzamanlı Reconcile sayısıdır. Varsayılan olarak 1'dir.
	MaxConcurrentReconciles int

	// CacheSyncTimeout, önbelleklerin senkronize edilmesini beklemek için ayarlanan zaman sınırını ifade eder.
	// Ayarlanmazsa varsayılan olarak 2 dakika olur.
	CacheSyncTimeout time.Duration

	// RecoverPanic, reconcile tarafından neden olunan paniğin kurtarılıp kurtarılmayacağını belirtir.
	// Ayarlanmamışsa, Yöneticiden Controller.RecoverPanic ayarına varsayılan olarak ayarlanır.
	// Yöneticiden Controller.RecoverPanic ayarı da ayarlanmamışsa varsayılan olarak true olur.
	RecoverPanic *bool

	// NeedLeaderElection, denetleyicinin lider seçimi kullanması gerekip gerekmediğini belirtir.
	// Varsayılan olarak true'dur, bu da denetleyicinin lider seçimi kullanacağı anlamına gelir.
	NeedLeaderElection *bool

	// Reconciler bir nesneyi reconcile eder
	Reconciler reconcile.TypedReconciler[request]

	// RateLimiter, isteklerin ne sıklıkta sıraya alınabileceğini sınırlamak için kullanılır.
	// Varsayılan olarak, hem genel hem de öğe başına oran sınırlaması olan MaxOfRateLimiter'dır.
	// Genel bir jeton kovasıdır ve öğe başına üstel bir sınırlamadır.
	RateLimiter workqueue.TypedRateLimiter[request]

	// NewQueue, denetleyici başlatılmaya hazır olduğunda bu denetleyici için kuyruğu oluşturur.
	// NewQueue ile özel bir kuyruk uygulaması kullanılabilir, örneğin, nesnelerin hangi öncelik/sıra ile reconcile edileceğini önceliklendirmek için bir öncelik kuyruğu.
	// Bu bir fonksiyondur çünkü standart Kubernetes iş kuyrukları hemen kendilerini başlatır, bu da birisi controller.New'u tekrar tekrar çağırırsa goroutine sızıntılarına yol açar.
	// NewQueue fonksiyonu, denetleyici adını ve (gerekirse varsayılan) RateLimiter seçeneğini alır.
	// NewQueue varsayılan olarak NewRateLimitingQueueWithConfig'tir.
	//
	// NOT: DÜŞÜK SEVİYE PRİMİTİF!
	// Sadece ne yaptığınızı biliyorsanız özel bir NewQueue kullanın.
	NewQueue func(controllerName string, rateLimiter workqueue.TypedRateLimiter[request]) workqueue.TypedRateLimitingInterface[request]

	// LogConstructor, bu denetleyici için kullanılan bir logger oluşturmak ve her reconcile işlemine context alanı aracılığıyla geçirmek için kullanılır.
	LogConstructor func(request *request) logr.Logger
}

// Controller bir Kubernetes API'sini uygular. Bir Controller, source.Sources'dan gelen reconcile.Request'leri besleyen bir iş kuyruğunu yönetir.
// İş, sıraya alınan her öğe için reconcile.Reconciler aracılığıyla gerçekleştirilir.
// İş tipik olarak, sistem durumunu nesne Spec'inde belirtilen durumla eşleşecek şekilde yapmak için Kubernetes nesnelerini okur ve yazar.
type Controller = TypedController[reconcile.Request]

// TypedController bir API uygular.
type TypedController[request comparable] interface {
	// Reconciler, Namespace/Name ile bir nesneyi reconcile etmek için çağrılır
	reconcile.TypedReconciler[request]

	// Watch, sağlanan Kaynağı izler.
	Watch(src source.TypedSource[request]) error

	// Start, denetleyiciyi başlatır. Start, context kapatılana veya bir
	// denetleyici başlatma hatası olana kadar bloklar.
	Start(ctx context.Context) error

	// GetLogger, temel bilgilerle önceden doldurulmuş bu denetleyici logger'ını döndürür.
	GetLogger() logr.Logger
}

// New, Yöneticide kayıtlı yeni bir Denetleyici döndürür. Yönetici, Denetleyici başlatılmadan önce paylaşılan Önbelleklerin senkronize edildiğinden emin olacaktır.
//
// Ad benzersiz olmalıdır çünkü metriklerde ve günlüklerde denetleyiciyi tanımlamak için kullanılır.
func New(name string, mgr manager.Manager, options Options) (Controller, error) {
	return NewTyped(name, mgr, options)
}

// NewTyped, Yöneticide kayıtlı yeni bir yazılı denetleyici döndürür,
//
// Ad benzersiz olmalıdır çünkü metriklerde ve günlüklerde denetleyiciyi tanımlamak için kullanılır.
func NewTyped[request comparable](name string, mgr manager.Manager, options TypedOptions[request]) (TypedController[request], error) {
	c, err := NewTypedUnmanaged(name, mgr, options)
	if err != nil {
		return nil, err
	}

	// Denetleyiciyi Yönetici bileşenleri olarak ekleyin
	return c, mgr.Add(c)
}

// NewUnmanaged, yöneticiyi eklemeden yeni bir denetleyici döndürür.
// Çağıran, döndürülen denetleyiciyi başlatmaktan sorumludur.
//
// Ad benzersiz olmalıdır çünkü metriklerde ve günlüklerde denetleyiciyi tanımlamak için kullanılır.
func NewUnmanaged(name string, mgr manager.Manager, options Options) (Controller, error) {
	return NewTypedUnmanaged(name, mgr, options)
}

// NewTypedUnmanaged, yöneticiyi eklemeden yeni bir yazılı denetleyici döndürür.
//
// Ad benzersiz olmalıdır çünkü metriklerde ve günlüklerde denetleyiciyi tanımlamak için kullanılır.
func NewTypedUnmanaged[request comparable](name string, mgr manager.Manager, options TypedOptions[request]) (TypedController[request], error) {
	if options.Reconciler == nil {
		return nil, fmt.Errorf("Reconciler belirtilmelidir")
	}

	if len(name) == 0 {
		return nil, fmt.Errorf("Denetleyici için Ad belirtilmelidir")
	}

	if options.SkipNameValidation == nil {
		options.SkipNameValidation = mgr.GetControllerOptions().SkipNameValidation
	}

	if options.SkipNameValidation == nil || !*options.SkipNameValidation {
		if err := checkName(name); err != nil {
			return nil, err
		}
	}

	if options.LogConstructor == nil {
		log := mgr.GetLogger().WithValues(
			"controller", name,
		)
		options.LogConstructor = func(in *request) logr.Logger {
			log := log
			if req, ok := any(in).(*reconcile.Request); ok && req != nil {
				log = log.WithValues(
					"object", klog.KRef(req.Namespace, req.Name),
					"namespace", req.Namespace, "name", req.Name,
				)
			}
			return log
		}
	}

	if options.MaxConcurrentReconciles <= 0 {
		if mgr.GetControllerOptions().MaxConcurrentReconciles > 0 {
			options.MaxConcurrentReconciles = mgr.GetControllerOptions().MaxConcurrentReconciles
		} else {
			options.MaxConcurrentReconciles = 1
		}
	}

	if options.CacheSyncTimeout == 0 {
		if mgr.GetControllerOptions().CacheSyncTimeout != 0 {
			options.CacheSyncTimeout = mgr.GetControllerOptions().CacheSyncTimeout
		} else {
			options.CacheSyncTimeout = 2 * time.Minute
		}
	}

	if options.RateLimiter == nil {
		options.RateLimiter = workqueue.DefaultTypedControllerRateLimiter[request]()
	}

	if options.NewQueue == nil {
		options.NewQueue = func(controllerName string, rateLimiter workqueue.TypedRateLimiter[request]) workqueue.TypedRateLimitingInterface[request] {
			return workqueue.NewTypedRateLimitingQueueWithConfig(rateLimiter, workqueue.TypedRateLimitingQueueConfig[request]{
				Name: controllerName,
			})
		}
	}

	if options.RecoverPanic == nil {
		options.RecoverPanic = mgr.GetControllerOptions().RecoverPanic
	}

	if options.NeedLeaderElection == nil {
		options.NeedLeaderElection = mgr.GetControllerOptions().NeedLeaderElection
	}

	// Bağımlılıkları ayarlanmış denetleyici oluştur
	return &controller.Controller[request]{
		Do:                      options.Reconciler,
		RateLimiter:             options.RateLimiter,
		NewQueue:                options.NewQueue,
		MaxConcurrentReconciles: options.MaxConcurrentReconciles,
		CacheSyncTimeout:        options.CacheSyncTimeout,
		Name:                    name,
		LogConstructor:          options.LogConstructor,
		RecoverPanic:            options.RecoverPanic,
		LeaderElected:           options.NeedLeaderElection,
	}, nil
}

// ReconcileIDFromContext, geçerli context'ten reconcileID'yi alır.
var ReconcileIDFromContext = controller.ReconcileIDFromContext
