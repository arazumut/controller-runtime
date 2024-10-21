Move cluster-specific code out of the manager
===================

## Motivasyon

Bugün, controller-runtime kullanarak birden fazla kümede çalışan denetleyiciler oluşturmak zaten mümkündür. Ancak, bu belgelenmemiştir ve doğrudan anlaşılır değildir, kullanıcıların bunu nasıl yapacaklarını anlamak için uygulama detaylarına bakmalarını gerektirir.

## Hedefler

* Birden fazla kümede çalışan denetleyiciler oluşturmanın kolay keşfedilebilir bir yolunu sağlamak
* `Runnables` yönetimini "kubeconfig gerektiren şeylerin" oluşturulmasından ayırmak
* Sadece bir kümede çalışan denetleyiciler oluşturan kullanıcılar için değişiklikler yapmamak

## Hedef Dışı

## Öneri

Şu anda, `./pkg/manager.Manager` iki amaca hizmet etmektedir:

* Denetleyicileri/diğer çalıştırılabilirleri çalıştırmak ve yaşam döngülerini yönetmek
* Kubernetes kümesiyle etkileşim kurmak için çeşitli şeyler kurmak, örneğin bir `Client` ve bir `Cache`

Bu, tek bir küme ile konuşan denetleyiciler oluştururken çok iyi çalışır, ancak bazı kullanım durumları birden fazla küme ile etkileşime giren denetleyiciler gerektirir. Bu çoklu küme kullanım durumu bugün çok gariptir, çünkü her küme için bir yönetici oluşturmayı ve tüm sonraki yöneticileri ilkine eklemeyi gerektirir.

Bu belge, tüm küme özel kodunu yöneticiden çıkarıp yeni bir paket ve arayüze taşımayı ve ardından bu arayüzü yöneticiye gömmeyi önerir. Bu, tek küme durumları için kullanımı aynı tutmayı ve bu değişikliği geriye dönük uyumlu bir şekilde tanıtmayı sağlar.

Ayrıca, yönetici tüm önbellekleri diğer `runnables` başlamadan önce başlatacak şekilde genişletilir.

Yeni `Cluster` arayüzü şu şekilde görünecektir:

```go
type Cluster interface {
	// SetFields, inject arayüzünü uygulayan bir nesne üzerinde küme özel bağımlılıkları ayarlayacaktır,
	// özellikle inject.Client, inject.Cache, inject.Scheme, inject.Config ve inject.APIReader
	SetFields(interface{}) error

	// GetConfig, başlatılmış bir Config döndürür
	GetConfig() *rest.Config

	// GetClient, Config ile yapılandırılmış bir istemci döndürür. Bu istemci
	// tam anlamıyla "doğrudan" bir istemci olmayabilir -- örneğin bir önbellekten okuyabilir.
	// Varsayılan uygulamanın nasıl çalıştığı hakkında daha fazla bilgi için Options.NewClient'a bakın.
	GetClient() client.Client

	// GetFieldIndexer, istemci ile yapılandırılmış bir client.FieldIndexer döndürür
	GetFieldIndexer() client.FieldIndexer

	// GetCache, bir cache.Cache döndürür
	GetCache() cache.Cache

	// GetEventRecorderFor, sağlanan ad için yeni bir EventRecorder döndürür
	GetEventRecorderFor(name string) record.EventRecorder

	// GetRESTMapper, bir RESTMapper döndürür
	GetRESTMapper() meta.RESTMapper

	// GetAPIReader, API sunucusunu kullanacak şekilde yapılandırılmış bir okuyucu döndürür.
	// Bu, nadiren ve yalnızca istemci kullanım durumunuza uymadığında kullanılmalıdır.
	GetAPIReader() client.Reader

	// GetScheme, başlatılmış bir Scheme döndürür
	GetScheme() *runtime.Scheme

	// Start, Cluster'a bağlantıyı başlatır
	Start(<-chan struct{}) error
}
```

Ve mevcut `Manager` arayüzü şu şekilde değişecektir:

```go
type Manager interface {
	// Cluster, bir kümeye bağlanmak için nesneleri tutar
	cluster.Cluster

	// Add, bileşen üzerinde istenen bağımlılıkları ayarlayacak ve Start çağrıldığında bileşenin
	// başlatılmasına neden olacaktır. Add, argüman için inject arayüzünü uygulayan herhangi bir bağımlılığı enjekte edecektir - örneğin inject.Client.
	// Bir Runnable, LeaderElectionRunnable arayüzünü uygulayıp uygulamadığına bağlı olarak, bir Runnable
	// lider seçim modunda (her zaman çalışan) veya lider seçim modunda (lider seçim etkinleştirilmişse yönetilen) çalıştırılabilir.
	Add(Runnable) error

	// Elected, bu yönetici bir grup yöneticinin lideri seçildiğinde kapatılır,
	// ya bir lider seçimi kazandığı için ya da lider seçimi yapılandırılmadığı için.
	Elected() <-chan struct{}

	// SetFields, inject arayüzünü uygulayan bir nesne üzerinde herhangi bir bağımlılığı ayarlayacaktır - örneğin inject.Client.
	SetFields(interface{}) error

	// AddMetricsExtraHandler, metrikleri sunan http sunucusunda path üzerinde ek bir işleyici ekler.
	// Örneğin pprof gibi bazı tanılama uç noktalarını kaydetmek yararlı olabilir. Bu uç noktaların hassas olduğu ve
	// genel olarak açığa çıkarılmaması gerektiği unutulmamalıdır.
	// Burada sunulan basit path -> işleyici eşlemesi yeterli değilse, yeni bir http sunucusu/dinleyici
	// Add yöntemi aracılığıyla yöneticiye Runnable olarak eklenmelidir.
	AddMetricsExtraHandler(path string, handler http.Handler) error

	// AddHealthzCheck, Healthz denetleyicisi eklemenizi sağlar
	AddHealthzCheck(name string, check healthz.Checker) error

	// AddReadyzCheck, Readyz denetleyicisi eklemenizi sağlar
	AddReadyzCheck(name string, check healthz.Checker) error

	// Start, tüm kayıtlı Denetleyicileri başlatır ve Stop kanalı kapanana kadar bloklar.
	// Herhangi bir denetleyiciyi başlatırken bir hata varsa bir hata döndürür.
	// Eğer Lider Seçimi kullanılıyorsa, bu döndükten hemen sonra ikili dosya kapatılmalıdır,
	// aksi takdirde lider seçim gerektiren bileşenler lider kilidi kaybedildikten sonra çalışmaya devam edebilir.
	Start(<-chan struct{}) error

	// GetWebhookServer, bir webhook.Server döndürür
	GetWebhookServer() *webhook.Server
}
```

Ayrıca, başlangıç sırasında, `Manager` tüm önbellekleri diğer şeylerden önce başlatabilmek için `Cluster`ları bulmak için tür doğrulaması kullanacaktır:

```go
type HasCaches interface {
  GetCache()
}
if getter, hasCaches := runnable.(HasCaches); hasCaches {
	m.caches = append(m.caches, getter())
}
```

```go
for idx := range cm.caches {
	go func(idx int) {cm.caches[idx].Start(cm.internalStop)}
}

for _, cache := range cm.caches {
	cache.WaitForCacheSync(cm.internalStop)
}

// Tüm diğer çalıştırılabilirleri başlat
```

## Örnek

Aşağıda, `referenceCluster`da bulunan her bir sır için `mirrorCluster`da aynı ada sahip bir sır yoksa bir sır oluşturacak bir `reconciler` örneği bulunmaktadır. Örneği kısa tutmak için, sırların içeriğini karşılaştırmayacaktır.

```go
type secretMirrorReconciler struct {
	referenceClusterClient, mirrorClusterClient client.Client
}

func (r *secretMirrorReconciler) Reconcile(r reconcile.Request)(reconcile.Result, error){
	s := &corev1.Secret{}
	if err := r.referenceClusterClient.Get(context.TODO(), r.NamespacedName, s); err != nil {
		if kerrors.IsNotFound{ return reconcile.Result{}, nil }
		return reconcile.Result, err
	}

	if err := r.mirrorClusterClient.Get(context.TODO(), r.NamespacedName, &corev1.Secret); err != nil {
		if !kerrors.IsNotFound(err) {
			return reconcile.Result{}, err
		}

		mirrorSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Namespace: s.Namespace, Name: s.Name},
			Data: s.Data,
		}
		return reconcile.Result{}, r.mirrorClusterClient.Create(context.TODO(), mirrorSecret)
	}

	return nil
}

func NewSecretMirrorReconciler(mgr manager.Manager, mirrorCluster cluster.Cluster) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Referans kümedeki Sırları izle
		For(&corev1.Secret{}).
		// Aynalama kümesindeki Sırları izle
		Watches(
			source.NewKindWithCache(&corev1.Secret{}, mirrorCluster.GetCache()),
			&handler.EnqueueRequestForObject{},
		).
		Complete(&secretMirrorReconciler{
			referenceClusterClient: mgr.GetClient(),
			mirrorClusterClient:    mirrorCluster.GetClient(),
		})
	}
}

func main(){

	mgr, err := manager.New( cfg1, manager.Options{})
	if err != nil {
		panic(err)
	}

	mirrorCluster, err := cluster.New(cfg2)
	if err != nil {
		panic(err)
	}

	if err := mgr.Add(mirrorCluster); err != nil {
		panic(err)
	}

	if err := NewSecretMirrorReconciler(mgr, mirrorCluster); err != nil {
		panic(err)
	}

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		panic(err)
	}
}
```
