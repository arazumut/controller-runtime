/*
2020 Kubernetes Yazarları tarafından oluşturulmuştur.

Apache Lisansı, Sürüm 2.0 ("Lisans") kapsamında lisanslanmıştır;
bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izinle gerekli olmadıkça,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN.
Lisans kapsamındaki izinler ve sınırlamalar hakkında daha fazla bilgi için
Lisans'a bakınız.
*/

package client

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NewNamespacedClient mevcut bir istemciyi belirli bir ad alanı değeri zorlayarak sarmalar.
// Bu istemciyi kullanan tüm fonksiyonlar burada belirtilen aynı ad alanına sahip olacaktır.
func NewNamespacedClient(c Client, ns string) Client {
	return &namespacedClient{
		client:    c,
		namespace: ns,
	}
}

var _ Client = &namespacedClient{}

// namespacedClient, belirli bir ad alanı değerini zorlamak için başka bir İstemciyi saran bir İstemcidir.
type namespacedClient struct {
	namespace string
	client    Client
}

// Scheme bu istemcinin kullandığı şemayı döndürür.
func (n *namespacedClient) Scheme() *runtime.Scheme {
	return n.client.Scheme()
}

// RESTMapper bu istemcinin kullandığı şemayı döndürür.
func (n *namespacedClient) RESTMapper() meta.RESTMapper {
	return n.client.RESTMapper()
}

// GroupVersionKindFor verilen nesne için GroupVersionKind döndürür.
func (n *namespacedClient) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return n.client.GroupVersionKindFor(obj)
}

// IsObjectNamespaced nesnenin GroupVersionKind'inin ad alanına sahip olup olmadığını döndürür.
func (n *namespacedClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return n.client.IsObjectNamespaced(obj)
}

// Create client.Client'i uygular.
func (n *namespacedClient) Create(ctx context.Context, obj Object, opts ...CreateOption) error {
	isNamespaceScoped, err := n.IsObjectNamespaced(obj)
	if err != nil {
		return fmt.Errorf("nesnenin kapsamını bulma hatası: %w", err)
	}

	objectNamespace := obj.GetNamespace()
	if objectNamespace != n.namespace && objectNamespace != "" {
		return fmt.Errorf("nesnenin %s ad alanı, istemcideki %s ad alanı ile eşleşmiyor", objectNamespace, n.namespace)
	}

	if isNamespaceScoped && objectNamespace == "" {
		obj.SetNamespace(n.namespace)
	}
	return n.client.Create(ctx, obj, opts...)
}

// Update client.Client'i uygular.
func (n *namespacedClient) Update(ctx context.Context, obj Object, opts ...UpdateOption) error {
	isNamespaceScoped, err := n.IsObjectNamespaced(obj)
	if err != nil {
		return fmt.Errorf("nesnenin kapsamını bulma hatası: %w", err)
	}

	objectNamespace := obj.GetNamespace()
	if objectNamespace != n.namespace && objectNamespace != "" {
		return fmt.Errorf("nesnenin %s ad alanı, istemcideki %s ad alanı ile eşleşmiyor", objectNamespace, n.namespace)
	}

	if isNamespaceScoped && objectNamespace == "" {
		obj.SetNamespace(n.namespace)
	}
	return n.client.Update(ctx, obj, opts...)
}

// Delete client.Client'i uygular.
func (n *namespacedClient) Delete(ctx context.Context, obj Object, opts ...DeleteOption) error {
	isNamespaceScoped, err := n.IsObjectNamespaced(obj)
	if err != nil {
		return fmt.Errorf("nesnenin kapsamını bulma hatası: %w", err)
	}

	objectNamespace := obj.GetNamespace()
	if objectNamespace != n.namespace && objectNamespace != "" {
		return fmt.Errorf("nesnenin %s ad alanı, istemcideki %s ad alanı ile eşleşmiyor", objectNamespace, n.namespace)
	}

	if isNamespaceScoped && objectNamespace == "" {
		obj.SetNamespace(n.namespace)
	}
	return n.client.Delete(ctx, obj, opts...)
}

// DeleteAllOf client.Client'i uygular.
func (n *namespacedClient) DeleteAllOf(ctx context.Context, obj Object, opts ...DeleteAllOfOption) error {
	isNamespaceScoped, err := n.IsObjectNamespaced(obj)
	if err != nil {
		return fmt.Errorf("nesnenin kapsamını bulma hatası: %w", err)
	}

	if isNamespaceScoped {
		opts = append(opts, InNamespace(n.namespace))
	}
	return n.client.DeleteAllOf(ctx, obj, opts...)
}

// Patch client.Client'i uygular.
func (n *namespacedClient) Patch(ctx context.Context, obj Object, patch Patch, opts ...PatchOption) error {
	isNamespaceScoped, err := n.IsObjectNamespaced(obj)
	if err != nil {
		return fmt.Errorf("nesnenin kapsamını bulma hatası: %w", err)
	}

	objectNamespace := obj.GetNamespace()
	if objectNamespace != n.namespace && objectNamespace != "" {
		return fmt.Errorf("nesnenin %s ad alanı, istemcideki %s ad alanı ile eşleşmiyor", objectNamespace, n.namespace)
	}

	if isNamespaceScoped && objectNamespace == "" {
		obj.SetNamespace(n.namespace)
	}
	return n.client.Patch(ctx, obj, patch, opts...)
}

// Get client.Client'i uygular.
func (n *namespacedClient) Get(ctx context.Context, key ObjectKey, obj Object, opts ...GetOption) error {
	isNamespaceScoped, err := n.IsObjectNamespaced(obj)
	if err != nil {
		return fmt.Errorf("nesnenin kapsamını bulma hatası: %w", err)
	}
	if isNamespaceScoped {
		if key.Namespace != "" && key.Namespace != n.namespace {
			return fmt.Errorf("nesne için sağlanan %s ad alanı, istemcideki %s ad alanı ile eşleşmiyor", key.Namespace, n.namespace)
		}
		key.Namespace = n.namespace
	}
	return n.client.Get(ctx, key, obj, opts...)
}

// List client.Client'i uygular.
func (n *namespacedClient) List(ctx context.Context, obj ObjectList, opts ...ListOption) error {
	if n.namespace != "" {
		opts = append(opts, InNamespace(n.namespace))
	}
	return n.client.List(ctx, obj, opts...)
}

// Status client.StatusClient'i uygular.
func (n *namespacedClient) Status() SubResourceWriter {
	return n.SubResource("status")
}

// SubResource client.SubResourceClient'i uygular.
func (n *namespacedClient) SubResource(subResource string) SubResourceClient {
	return &namespacedClientSubResourceClient{client: n.client.SubResource(subResource), namespace: n.namespace, namespacedclient: n}
}

// namespacedClientSubResourceClient'in client.SubResourceClient'i uyguladığından emin olun.
var _ SubResourceClient = &namespacedClientSubResourceClient{}

type namespacedClientSubResourceClient struct {
	client           SubResourceClient
	namespace        string
	namespacedclient Client
}

func (nsw *namespacedClientSubResourceClient) Get(ctx context.Context, obj, subResource Object, opts ...SubResourceGetOption) error {
	isNamespaceScoped, err := nsw.namespacedclient.IsObjectNamespaced(obj)
	if err != nil {
		return fmt.Errorf("nesnenin kapsamını bulma hatası: %w", err)
	}

	objectNamespace := obj.GetNamespace()
	if objectNamespace != nsw.namespace && objectNamespace != "" {
		return fmt.Errorf("nesnenin %s ad alanı, istemcideki %s ad alanı ile eşleşmiyor", objectNamespace, nsw.namespace)
	}

	if isNamespaceScoped && objectNamespace == "" {
		obj.SetNamespace(nsw.namespace)
	}

	return nsw.client.Get(ctx, obj, subResource, opts...)
}

func (nsw *namespacedClientSubResourceClient) Create(ctx context.Context, obj, subResource Object, opts ...SubResourceCreateOption) error {
	isNamespaceScoped, err := nsw.namespacedclient.IsObjectNamespaced(obj)
	if err != nil {
		return fmt.Errorf("nesnenin kapsamını bulma hatası: %w", err)
	}

	objectNamespace := obj.GetNamespace()
	if objectNamespace != nsw.namespace && objectNamespace != "" {
		return fmt.Errorf("nesnenin %s ad alanı, istemcideki %s ad alanı ile eşleşmiyor", objectNamespace, nsw.namespace)
	}

	if isNamespaceScoped && objectNamespace == "" {
		obj.SetNamespace(nsw.namespace)
	}

	return nsw.client.Create(ctx, obj, subResource, opts...)
}

// Update client.SubResourceWriter'i uygular.
func (nsw *namespacedClientSubResourceClient) Update(ctx context.Context, obj Object, opts ...SubResourceUpdateOption) error {
	isNamespaceScoped, err := nsw.namespacedclient.IsObjectNamespaced(obj)
	if err != nil {
		return fmt.Errorf("nesnenin kapsamını bulma hatası: %w", err)
	}

	objectNamespace := obj.GetNamespace()
	if objectNamespace != nsw.namespace && objectNamespace != "" {
		return fmt.Errorf("nesnenin %s ad alanı, istemcideki %s ad alanı ile eşleşmiyor", objectNamespace, nsw.namespace)
	}

	if isNamespaceScoped && objectNamespace == "" {
		obj.SetNamespace(nsw.namespace)
	}
	return nsw.client.Update(ctx, obj, opts...)
}

// Patch client.SubResourceWriter'i uygular.
func (nsw *namespacedClientSubResourceClient) Patch(ctx context.Context, obj Object, patch Patch, opts ...SubResourcePatchOption) error {
	isNamespaceScoped, err := nsw.namespacedclient.IsObjectNamespaced(obj)
	if err != nil {
		return fmt.Errorf("nesnenin kapsamını bulma hatası: %w", err)
	}

	objectNamespace := obj.GetNamespace()
	if objectNamespace != nsw.namespace && objectNamespace != "" {
		return fmt.Errorf("nesnenin %s ad alanı, istemcideki %s ad alanı ile eşleşmiyor", objectNamespace, nsw.namespace)
	}

	if isNamespaceScoped && objectNamespace == "" {
		obj.SetNamespace(nsw.namespace)
	}
	return nsw.client.Patch(ctx, obj, patch, opts...)
}
