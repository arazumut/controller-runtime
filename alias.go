/*
Kubernetes Yazarları 2018.

Apache Lisansı, Sürüm 2.0 ("Lisans") kapsamında lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa kapsamında gerekli olmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisansa bakın.
*/

package controllerruntime

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Builder, bir Uygulama ControllerManagedBy (örneğin Operatör) oluşturur ve başlatmak için bir manager.Manager döner.
type Builder = builder.Builder

// Request, bir Kubernetes nesnesini uzlaştırmak için gerekli bilgileri içerir. Bu, nesneyi benzersiz şekilde tanımlamak için
// gerekli bilgileri içerir - Adı ve Ad Alanı. Belirli bir Olay veya nesne içeriği hakkında bilgi içermez.
type Request = reconcile.Request

// Result, bir Reconciler çağrısının sonucunu içerir.
type Result = reconcile.Result

// Manager, Önbellekler ve İstemciler gibi paylaşılan bağımlılıkları başlatır ve bunları Runnables'a sağlar.
// Bir Manager, Kontrolörler oluşturmak için gereklidir.
type Manager = manager.Manager

// Options, yeni bir Manager oluşturmak için argümanlardır.
type Options = manager.Options

// SchemeBuilder, go türlerini Kubernetes GroupVersionKinds ile eşlemek için yeni bir Şema oluşturur.
type SchemeBuilder = scheme.Builder

// GroupVersion, API'yi benzersiz şekilde tanımlayan "grup" ve "sürüm" içerir.
type GroupVersion = schema.GroupVersion

// GroupResource, bir Grup ve bir Kaynak belirtir, ancak bir sürümü zorlamaz. Bu, kısmen geçerli türler olmadan
// kavramları arama aşamalarında tanımlamak için kullanışlıdır.
type GroupResource = schema.GroupResource

// TypeMeta, bir API yanıtında veya isteğinde bireysel bir nesneyi tanımlar
// ve nesnenin türünü ve API şema sürümünü temsil eden dizeler içerir.
// Sürümlenen veya kalıcı hale getirilen yapılar TypeMeta'yı içermelidir.
//
// +k8s:deepcopy-gen=false
type TypeMeta = metav1.TypeMeta

// ObjectMeta, tüm kalıcı kaynakların sahip olması gereken meta verileri içerir, bu da kullanıcıların oluşturması gereken tüm nesneleri içerir.
type ObjectMeta = metav1.ObjectMeta

var (
	// RegisterFlags, belirtilen FlagSet'e bayrak değişkenlerini kaydeder, eğer zaten kayıtlı değilse.
	// Varsayılan komut satırı FlagSet'i kullanır, eğer sağlanmamışsa. Şu anda yalnızca kubeconfig bayrağını kaydeder.
	RegisterFlags = config.RegisterFlags

	// GetConfigOrDie, bir Kubernetes apiserver ile konuşmak için bir *rest.Config oluşturur.
	// Eğer --kubeconfig ayarlanmışsa, o konumdaki kubeconfig dosyasını kullanır. Aksi takdirde, küme içinde çalıştığını varsayar
	// ve küme tarafından sağlanan kubeconfig'i kullanır.
	//
	// rest.Config oluştururken bir hata oluşursa bir hata günlüğü kaydeder ve çıkar.
	GetConfigOrDie = config.GetConfigOrDie

	// GetConfig, bir Kubernetes apiserver ile konuşmak için bir *rest.Config oluşturur.
	// Eğer --kubeconfig ayarlanmışsa, o konumdaki kubeconfig dosyasını kullanır. Aksi takdirde, küme içinde çalıştığını varsayar
	// ve küme tarafından sağlanan kubeconfig'i kullanır.
	//
	// Config önceliği
	//
	// * --kubeconfig bayrağı bir dosyaya işaret ediyorsa
	//
	// * KUBECONFIG ortam değişkeni bir dosyaya işaret ediyorsa
	//
	// * Küme içinde çalışıyorsa küme içi yapılandırma
	//
	// * $HOME/.kube/config varsa.
	GetConfig = config.GetConfig

	// NewControllerManagedBy, sağlanan Manager tarafından başlatılacak yeni bir kontrolör oluşturucu döner.
	NewControllerManagedBy = builder.ControllerManagedBy

	// NewWebhookManagedBy, sağlanan Manager tarafından başlatılacak yeni bir webhook oluşturucu döner.
	NewWebhookManagedBy = builder.WebhookManagedBy

	// NewManager, Kontrolörler oluşturmak için yeni bir Manager döner.
	// Verilen yapılandırmadaki ContentType ayarlanmamışsa, Kubernetes'in tüm yerleşik kaynakları için "application/vnd.kubernetes.protobuf"
	// ve diğer türler için "application/json" kullanılacaktır, CRD kaynakları dahil.
	NewManager = manager.New

	// CreateOrUpdate, verilen obj nesnesini Kubernetes kümesinde oluşturur veya günceller.
	// Nesnenin istenen durumu, geçirilen ReconcileFn kullanılarak mevcut durumla uzlaştırılmalıdır.
	// obj, sunucu tarafından döndürülen içerikle güncellenebilmesi için bir yapı işaretçisi olmalıdır.
	//
	// Gerçekleştirilen işlemi ve bir hatayı döner.
	CreateOrUpdate = controllerutil.CreateOrUpdate

	// SetControllerReference, owner'ı owned üzerinde bir Kontrolör SahiplikReferansı olarak ayarlar.
	// Bu, owned nesnesinin çöp toplanması ve owned üzerinde değişiklikler olduğunda owner nesnesinin uzlaştırılması için kullanılır
	// (bir Watch + EnqueueRequestForOwner ile).
	// Yalnızca bir SahiplikReferansı bir kontrolör olabilir, bu nedenle başka bir SahiplikReferansı varsa
	// Kontrolör bayrağı ayarlanmışsa bir hata döner.
	SetControllerReference = controllerutil.SetControllerReference

	// SetupSignalHandler, SIGTERM ve SIGINT için kayıt yapar. Bir context döner
	// bu sinyallerden biri alındığında iptal edilir. İkinci bir sinyal alınırsa, program
	// çıkış kodu 1 ile sonlandırılır.
	SetupSignalHandler = signals.SetupSignalHandler

	// Log, controller-runtime tarafından kullanılan temel logger'dır. Başka bir logr.Logger'a devreder.
	// Gerçek bir günlükleme almak için SetLogger'ı çağırmalısınız.
	Log = log.Log

	// LoggerFrom, context.Context'ten önceden tanımlanmış değerlerle bir logger döner.
	// Logger, kontrolörlerle kullanıldığında, uzlaştırılan nesne hakkında temel bilgiler içermesi beklenebilir:
	// - For(...) nesnesi kontrolör oluştururken geçirildiğinde `uzlaştırıcı grup` ve `uzlaştırıcı tür`.
	// - Uzlaştırma isteğinden `ad` ve `ad alanı`.
	//
	// Bu, Reconciler arayüzünü karşılayan bir yapıdaki context ile kullanılmak üzere tasarlanmıştır.
	LoggerFrom = log.FromContext

	// LoggerInto, bir context alır ve logger'ı anahtarlarından biri olarak ayarlar.
	//
	// Bu, uzlaştırıcılarda bir context içindeki logger'ı ek değerlerle zenginleştirmek için tasarlanmıştır.
	LoggerInto = log.IntoContext

	// SetLogger, tüm ertelenmiş Logger'lar için somut bir günlükleme uygulaması ayarlar.
	SetLogger = log.SetLogger
)
