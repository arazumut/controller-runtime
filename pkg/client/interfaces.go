/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa veya yazılı izin gereği olmadıkça,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMADAN.
Lisans kapsamındaki izinler ve sınırlamalar için Lisansa bakınız.
*/

package client

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

// ObjectKey bir Kubernetes Nesnesini tanımlar.
type ObjectKey = types.NamespacedName

// ObjectKeyFromObject, bir runtime.Object için ObjectKey döndürür.
func ObjectKeyFromObject(obj Object) ObjectKey {
	return ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}
}

// Patch, bir Kubernetes nesnesine uygulanabilecek bir yamadır.
type Patch interface {
	// Type, yamanın PatchType'ıdır.
	Type() types.PatchType
	// Data, yamayı temsil eden ham veridir.
	Data(obj Object) ([]byte, error)
}

// TODO(directxman12): get/delete seçenekleriyle başa çıkmanın mantıklı bir yolu var mı?

// Reader, Kubernetes nesnelerini nasıl okuyacağını ve listeleyeceğini bilir.
type Reader interface {
	// Get, verilen nesne anahtarı için Kubernetes Kümesinden bir nesne alır.
	// obj, sunucudan döndürülen yanıtla güncellenebilmesi için bir yapı işaretçisi olmalıdır.
	Get(ctx context.Context, key ObjectKey, obj Object, opts ...GetOption) error

	// List, verilen ad alanı ve liste seçenekleri için nesne listesini alır. Başarılı bir çağrıda,
	// list içindeki Items alanı sunucudan döndürülen sonuçla doldurulacaktır.
	List(ctx context.Context, list ObjectList, opts ...ListOption) error
}

// Writer, Kubernetes nesnelerini nasıl oluşturacağını, sileceğini ve güncelleyeceğini bilir.
type Writer interface {
	// Create, obj nesnesini Kubernetes kümesinde kaydeder. obj, sunucudan döndürülen içerikle güncellenebilmesi için bir yapı işaretçisi olmalıdır.
	Create(ctx context.Context, obj Object, opts ...CreateOption) error

	// Delete, verilen obj nesnesini Kubernetes kümesinden siler.
	Delete(ctx context.Context, obj Object, opts ...DeleteOption) error

	// Update, verilen obj nesnesini Kubernetes kümesinde günceller. obj, sunucudan döndürülen içerikle güncellenebilmesi için bir yapı işaretçisi olmalıdır.
	Update(ctx context.Context, obj Object, opts ...UpdateOption) error

	// Patch, verilen obj nesnesini Kubernetes kümesinde yamalar. obj, sunucudan döndürülen içerikle güncellenebilmesi için bir yapı işaretçisi olmalıdır.
	Patch(ctx context.Context, obj Object, patch Patch, opts ...PatchOption) error

	// DeleteAllOf, verilen seçeneklere uyan tüm nesneleri siler.
	DeleteAllOf(ctx context.Context, obj Object, opts ...DeleteAllOfOption) error
}

// StatusClient, Kubernetes nesneleri için durum alt kaynağını güncelleyebilen bir istemci oluşturmayı bilir.
type StatusClient interface {
	Status() SubResourceWriter
}

// SubResourceClientConstructor, Kubernetes nesneleri için alt kaynağı güncelleyebilen bir istemci oluşturmayı bilir.
type SubResourceClientConstructor interface {
	// SubResourceClientConstructor, adlandırılmış alt kaynak için bir alt kaynak istemcisi döndürür. Bilinen
	// yukarı akış alt kaynak kullanımları şunlardır:
	// - ServiceAccount token oluşturma:
	//     sa := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}}
	//     token := &authenticationv1.TokenRequest{}
	//     c.SubResourceClient("token").Create(ctx, sa, token)
	//
	// - Pod tahliye oluşturma:
	//     pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}}
	//     c.SubResourceClient("eviction").Create(ctx, pod, &policyv1.Eviction{})
	//
	// - Pod bağlama oluşturma:
	//     pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}}
	//     binding := &corev1.Binding{Target: corev1.ObjectReference{Name: "my-node"}}
	//     c.SubResourceClient("binding").Create(ctx, pod, binding)
	//
	// - CertificateSigningRequest onayı:
	//     csr := &certificatesv1.CertificateSigningRequest{
	//	     ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"},
	//       Status: certificatesv1.CertificateSigningRequestStatus{
	//         Conditions: []certificatesv1.[]CertificateSigningRequestCondition{{
	//           Type: certificatesv1.CertificateApproved,
	//           Status: corev1.ConditionTrue,
	//         }},
	//       },
	//     }
	//     c.SubResourceClient("approval").Update(ctx, csr)
	//
	// - Ölçek alma:
	//     dep := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}}
	//     scale := &autoscalingv1.Scale{}
	//     c.SubResourceClient("scale").Get(ctx, dep, scale)
	//
	// - Ölçek güncelleme:
	//     dep := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"}}
	//     scale := &autoscalingv1.Scale{Spec: autoscalingv1.ScaleSpec{Replicas: 2}}
	//     c.SubResourceClient("scale").Update(ctx, dep, client.WithSubResourceBody(scale))
	SubResource(subResource string) SubResourceClient
}

// StatusWriter, geriye dönük uyumluluk için tutulmuştur.
type StatusWriter = SubResourceWriter

// SubResourceReader, Alt Kaynakları nasıl okuyacağını bilir.
type SubResourceReader interface {
	Get(ctx context.Context, obj Object, subResource Object, opts ...SubResourceGetOption) error
}

// SubResourceWriter, bir Kubernetes nesnesinin alt kaynağını nasıl güncelleyeceğini bilir.
type SubResourceWriter interface {
	// Create, alt kaynak nesnesini Kubernetes kümesinde kaydeder. obj, sunucudan döndürülen içerikle güncellenebilmesi için bir yapı işaretçisi olmalıdır.
	Create(ctx context.Context, obj Object, subResource Object, opts ...SubResourceCreateOption) error

	// Update, verilen nesne için durum alt kaynağına karşılık gelen alanları günceller. obj, sunucudan döndürülen içerikle güncellenebilmesi için bir yapı işaretçisi olmalıdır.
	Update(ctx context.Context, obj Object, opts ...SubResourceUpdateOption) error

	// Patch, verilen nesnenin alt kaynağını yamalar. obj, sunucudan döndürülen içerikle güncellenebilmesi için bir yapı işaretçisi olmalıdır.
	Patch(ctx context.Context, obj Object, patch Patch, opts ...SubResourcePatchOption) error
}

// SubResourceClient, Kubernetes nesneleri üzerinde CRU işlemlerini nasıl gerçekleştireceğini bilir.
type SubResourceClient interface {
	SubResourceReader
	SubResourceWriter
}

// Client, Kubernetes nesneleri üzerinde CRUD işlemlerini nasıl gerçekleştireceğini bilir.
type Client interface {
	Reader
	Writer
	StatusClient
	SubResourceClientConstructor

	// Scheme, bu istemcinin kullandığı şemayı döndürür.
	Scheme() *runtime.Scheme
	// RESTMapper, bu istemcinin kullandığı rest'i döndürür.
	RESTMapper() meta.RESTMapper
	// GroupVersionKindFor, verilen nesne için GroupVersionKind'ı döndürür.
	GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error)
	// IsObjectNamespaced, nesnenin GroupVersionKind'ının ad alanına sahip olup olmadığını döndürür.
	IsObjectNamespaced(obj runtime.Object) (bool, error)
}

// WithWatch, normal İstemci tarafından desteklenen CRUD işlemlerinin üzerine İzleme desteği sağlar. Amacı, olayları beklemesi gereken CLI uygulamalarıdır.
type WithWatch interface {
	Client
	Watch(ctx context.Context, obj ObjectList, opts ...ListOption) (watch.Interface, error)
}

// IndexerFunc, bir nesneyi alıp bir dizi ad alanına sahip olmayan anahtara dönüştürmeyi bilir. Ad alanına sahip nesneler otomatik olarak ad alanına sahip ve ad alanına sahip olmayan varyantlar alır, bu nedenle anahtarlar ad alanını içermemelidir.
type IndexerFunc func(Object) []string

// FieldIndexer, belirli bir "alan" üzerinde nasıl indeksleme yapacağını bilir, böylece daha sonra bir alan seçici tarafından kullanılabilir.
type FieldIndexer interface {
	// IndexFields, verilen nesne türünde verilen alan adıyla bir indeks ekler
	// bu alan için değeri çıkarmak için verilen işlevi kullanarak. Kubernetes API sunucusuyla uyumluluk istiyorsanız, yalnızca bir anahtar döndürün ve yalnızca
	// API sunucusunun desteklediği alanları kullanın. Aksi takdirde, birden fazla anahtar döndürebilirsiniz ve alan seçicide "eşitlik" en az bir anahtarın değeri eşleştirmesi anlamına gelir.
	// FieldIndexer, ad alanı üzerinde indeksleme yapmayı ve tüm ad alanı sorgularını desteklemeyi otomatik olarak halleder.
	IndexField(ctx context.Context, obj Object, field string, extractValue IndexerFunc) error
}

// IgnoreNotFound, NotFound hatalarında nil döndürür.
// NotFound hatası veya nil olmayan diğer tüm değerler değiştirilmeden döndürülür.
func IgnoreNotFound(err error) error {
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

// IgnoreAlreadyExists, AlreadyExists hatalarında nil döndürür.
// AlreadyExists hatası veya nil olmayan diğer tüm değerler değiştirilmeden döndürülür.
func IgnoreAlreadyExists(err error) error {
	if apierrors.IsAlreadyExists(err) {
		return nil
	}

	return err
}
