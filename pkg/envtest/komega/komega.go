/*
2021 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
yetkiler ve sınırlamalar için Lisans'a bakınız.
*/

package komega

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// komega, sahte bir Kubernetes API'si içeren testler yazmak için bir dizi yardımcı işlevdir.
type komega struct {
	ctx    context.Context
	client client.Client
}

var _ Komega = &komega{}

// Yeni bir Komega örneği oluşturur ve verilen istemciyi kullanır.
func New(c client.Client) Komega {
	return &komega{
		client: c,
		ctx:    context.Background(),
	}
}

// Verilen bağlamı kullanan bir kopya döndürür.
func (k komega) WithContext(ctx context.Context) Komega {
	k.ctx = ctx
	return &k
}

// Bir kaynağı getiren ve oluşan hatayı döndüren bir işlev döndürür.
func (k *komega) Get(obj client.Object) func() error {
	key := types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}
	return func() error {
		return k.client.Get(k.ctx, key, obj)
	}
}

// Kaynakları listeleyen ve oluşan hatayı döndüren bir işlev döndürür.
func (k *komega) List(obj client.ObjectList, opts ...client.ListOption) func() error {
	return func() error {
		return k.client.List(k.ctx, obj, opts...)
	}
}

// Bir kaynağı getiren, sağlanan güncelleme işlevini uygulayan ve ardından kaynağı güncelleyen bir işlev döndürür.
func (k *komega) Update(obj client.Object, updateFunc func(), opts ...client.UpdateOption) func() error {
	key := types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}
	return func() error {
		err := k.client.Get(k.ctx, key, obj)
		if err != nil {
			return err
		}
		updateFunc()
		return k.client.Update(k.ctx, obj, opts...)
	}
}

// Bir kaynağı getiren, sağlanan güncelleme işlevini uygulayan ve ardından kaynağın durumunu güncelleyen bir işlev döndürür.
func (k *komega) UpdateStatus(obj client.Object, updateFunc func(), opts ...client.SubResourceUpdateOption) func() error {
	key := types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}
	return func() error {
		err := k.client.Get(k.ctx, key, obj)
		if err != nil {
			return err
		}
		updateFunc()
		return k.client.Status().Update(k.ctx, obj, opts...)
	}
}

// Bir kaynağı getiren ve nesneyi döndüren bir işlev döndürür.
func (k *komega) Object(obj client.Object) func() (client.Object, error) {
	key := types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}
	return func() (client.Object, error) {
		err := k.client.Get(k.ctx, key, obj)
		return obj, err
	}
}

// Bir kaynağı getiren ve nesne listesini döndüren bir işlev döndürür.
func (k *komega) ObjectList(obj client.ObjectList, opts ...client.ListOption) func() (client.ObjectList, error) {
	return func() (client.ObjectList, error) {
		err := k.client.List(k.ctx, obj, opts...)
		return obj, err
	}
}
