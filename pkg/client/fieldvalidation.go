/*
2024 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ OLMAKSIZIN, açık veya zımni olarak.
Lisans kapsamındaki izin ve sınırlamalar için Lisans'a bakınız.
*/

package client

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// WithFieldValidation, bir Client'i sarar ve bu istemciden gelen tüm yazma istekleri için
// varsayılan olarak alan doğrulamasını yapılandırır. Kullanıcılar, bireysel yazma istekleri
// için alan doğrulamasını geçersiz kılabilir.
func WithFieldValidation(c Client, validation FieldValidation) Client {
	return &clientWithFieldValidation{
		validation: validation,
		client:     c,
		Reader:     c,
	}
}

type clientWithFieldValidation struct {
	validation FieldValidation
	client     Client
	Reader
}

func (c *clientWithFieldValidation) Create(ctx context.Context, obj Object, opts ...CreateOption) error {
	return c.client.Create(ctx, obj, append([]CreateOption{c.validation}, opts...)...)
}

func (c *clientWithFieldValidation) Update(ctx context.Context, obj Object, opts ...UpdateOption) error {
	return c.client.Update(ctx, obj, append([]UpdateOption{c.validation}, opts...)...)
}

func (c *clientWithFieldValidation) Patch(ctx context.Context, obj Object, patch Patch, opts ...PatchOption) error {
	return c.client.Patch(ctx, obj, patch, append([]PatchOption{c.validation}, opts...)...)
}

func (c *clientWithFieldValidation) Delete(ctx context.Context, obj Object, opts ...DeleteOption) error {
	return c.client.Delete(ctx, obj, opts...)
}

func (c *clientWithFieldValidation) DeleteAllOf(ctx context.Context, obj Object, opts ...DeleteAllOfOption) error {
	return c.client.DeleteAllOf(ctx, obj, opts...)
}

func (c *clientWithFieldValidation) Scheme() *runtime.Scheme     { return c.client.Scheme() }
func (c *clientWithFieldValidation) RESTMapper() meta.RESTMapper { return c.client.RESTMapper() }
func (c *clientWithFieldValidation) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return c.client.GroupVersionKindFor(obj)
}

func (c *clientWithFieldValidation) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return c.client.IsObjectNamespaced(obj)
}

func (c *clientWithFieldValidation) Status() StatusWriter {
	return &subresourceClientWithFieldValidation{
		validation:        c.validation,
		subresourceWriter: c.client.Status(),
	}
}

func (c *clientWithFieldValidation) SubResource(subresource string) SubResourceClient {
	srClient := c.client.SubResource(subresource)
	return &subresourceClientWithFieldValidation{
		validation:        c.validation,
		subresourceWriter: srClient,
		SubResourceReader: srClient,
	}
}

type subresourceClientWithFieldValidation struct {
	validation        FieldValidation
	subresourceWriter SubResourceWriter
	SubResourceReader
}

func (c *subresourceClientWithFieldValidation) Create(ctx context.Context, obj Object, subresource Object, opts ...SubResourceCreateOption) error {
	return c.subresourceWriter.Create(ctx, obj, subresource, append([]SubResourceCreateOption{c.validation}, opts...)...)
}

func (c *subresourceClientWithFieldValidation) Update(ctx context.Context, obj Object, opts ...SubResourceUpdateOption) error {
	return c.subresourceWriter.Update(ctx, obj, append([]SubResourceUpdateOption{c.validation}, opts...)...)
}

func (c *subresourceClientWithFieldValidation) Patch(ctx context.Context, obj Object, patch Patch, opts ...SubResourcePatchOption) error {
	return c.subresourceWriter.Patch(ctx, obj, patch, append([]SubResourcePatchOption{c.validation}, opts...)...)
}
