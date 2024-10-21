/*
2020 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa gereği veya yazılı olarak kabul edilmediği sürece,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisansa bakınız.
*/

package client

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NewDryRunClient mevcut bir istemciyi sarar ve tüm değiştirici API çağrılarında
// DryRun modunu zorlar.
func NewDryRunClient(c Client) Client {
	return &dryRunClient{client: c}
}

var _ Client = &dryRunClient{}

// dryRunClient, DryRun modunu zorlamak için başka bir İstemciyi saran bir İstemcidir.
type dryRunClient struct {
	client Client
}

// Scheme bu istemcinin kullandığı şemayı döndürür.
func (c *dryRunClient) Scheme() *runtime.Scheme {
	return c.client.Scheme()
}

// RESTMapper bu istemcinin kullandığı rest mapper'ı döndürür.
func (c *dryRunClient) RESTMapper() meta.RESTMapper {
	return c.client.RESTMapper()
}

// GroupVersionKindFor verilen nesne için GroupVersionKind döndürür.
func (c *dryRunClient) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return c.client.GroupVersionKindFor(obj)
}

// IsObjectNamespaced nesnenin GroupVersionKind'inin ad alanına sahip olup olmadığını döndürür.
func (c *dryRunClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return c.client.IsObjectNamespaced(obj)
}

// Create client.Client'i uygular.
func (c *dryRunClient) Create(ctx context.Context, obj Object, opts ...CreateOption) error {
	return c.client.Create(ctx, obj, append(opts, DryRunAll)...)
}

// Update client.Client'i uygular.
func (c *dryRunClient) Update(ctx context.Context, obj Object, opts ...UpdateOption) error {
	return c.client.Update(ctx, obj, append(opts, DryRunAll)...)
}

// Delete client.Client'i uygular.
func (c *dryRunClient) Delete(ctx context.Context, obj Object, opts ...DeleteOption) error {
	return c.client.Delete(ctx, obj, append(opts, DryRunAll)...)
}

// DeleteAllOf client.Client'i uygular.
func (c *dryRunClient) DeleteAllOf(ctx context.Context, obj Object, opts ...DeleteAllOfOption) error {
	return c.client.DeleteAllOf(ctx, obj, append(opts, DryRunAll)...)
}

// Patch client.Client'i uygular.
func (c *dryRunClient) Patch(ctx context.Context, obj Object, patch Patch, opts ...PatchOption) error {
	return c.client.Patch(ctx, obj, patch, append(opts, DryRunAll)...)
}

// Get client.Client'i uygular.
func (c *dryRunClient) Get(ctx context.Context, key ObjectKey, obj Object, opts ...GetOption) error {
	return c.client.Get(ctx, key, obj, opts...)
}

// List client.Client'i uygular.
func (c *dryRunClient) List(ctx context.Context, obj ObjectList, opts ...ListOption) error {
	return c.client.List(ctx, obj, opts...)
}

// Status client.StatusClient'i uygular.
func (c *dryRunClient) Status() SubResourceWriter {
	return c.SubResource("status")
}

// SubResource client.SubResourceClient'i uygular.
func (c *dryRunClient) SubResource(altKaynak string) SubResourceClient {
	return &dryRunSubResourceClient{client: c.client.SubResource(altKaynak)}
}

// dryRunSubResourceWriter'ın client.SubResourceWriter'ı uyguladığından emin olun.
var _ SubResourceWriter = &dryRunSubResourceClient{}

// dryRunSubResourceClient, dryRun modunu zorlayarak durum alt kaynağını yazan client.SubResourceWriter'dır.
type dryRunSubResourceClient struct {
	client SubResourceClient
}

func (sw *dryRunSubResourceClient) Get(ctx context.Context, obj, altKaynak Object, opts ...SubResourceGetOption) error {
	return sw.client.Get(ctx, obj, altKaynak, opts...)
}

func (sw *dryRunSubResourceClient) Create(ctx context.Context, obj, altKaynak Object, opts ...SubResourceCreateOption) error {
	return sw.client.Create(ctx, obj, altKaynak, append(opts, DryRunAll)...)
}

// Update client.SubResourceWriter'ı uygular.
func (sw *dryRunSubResourceClient) Update(ctx context.Context, obj Object, opts ...SubResourceUpdateOption) error {
	return sw.client.Update(ctx, obj, append(opts, DryRunAll)...)
}

// Patch client.SubResourceWriter'ı uygular.
func (sw *dryRunSubResourceClient) Patch(ctx context.Context, obj Object, patch Patch, opts ...SubResourcePatchOption) error {
	return sw.client.Patch(ctx, obj, patch, append(opts, DryRunAll)...)
}
