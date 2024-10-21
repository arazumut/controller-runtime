/*
2024 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
bu yazılım Lisans kapsamında "OLDUĞU GİBİ" dağıtılmakta olup,
HERHANGİ BİR GARANTİ VERİLMEMEKTEDİR.
Lisans kapsamındaki izin ve sınırlamalar için Lisansa bakınız.
*/

package client

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// WithFieldOwner, bir Client'i sarar ve bu istemciden gelen tüm yazma isteklerine
// fieldOwner'ı alan yöneticisi olarak ekler. Bu istemcinin yöntemlerinde ek
// [FieldOwner] seçenekleri belirtilirse, burada belirtilen değer geçersiz kılınır.
func WithFieldOwner(c Client, fieldOwner string) Client {
	return &alanYöneticiliİstemci{
		sahip:  fieldOwner,
		c:      c,
		Reader: c,
	}
}

type alanYöneticiliİstemci struct {
	sahip string
	c     Client
	Reader
}

func (f *alanYöneticiliİstemci) Create(ctx context.Context, obj Object, opts ...CreateOption) error {
	return f.c.Create(ctx, obj, append([]CreateOption{FieldOwner(f.sahip)}, opts...)...)
}

func (f *alanYöneticiliİstemci) Update(ctx context.Context, obj Object, opts ...UpdateOption) error {
	return f.c.Update(ctx, obj, append([]UpdateOption{FieldOwner(f.sahip)}, opts...)...)
}

func (f *alanYöneticiliİstemci) Patch(ctx context.Context, obj Object, patch Patch, opts ...PatchOption) error {
	return f.c.Patch(ctx, obj, patch, append([]PatchOption{FieldOwner(f.sahip)}, opts...)...)
}

func (f *alanYöneticiliİstemci) Delete(ctx context.Context, obj Object, opts ...DeleteOption) error {
	return f.c.Delete(ctx, obj, opts...)
}

func (f *alanYöneticiliİstemci) DeleteAllOf(ctx context.Context, obj Object, opts ...DeleteAllOfOption) error {
	return f.c.DeleteAllOf(ctx, obj, opts...)
}

func (f *alanYöneticiliİstemci) Scheme() *runtime.Scheme     { return f.c.Scheme() }
func (f *alanYöneticiliİstemci) RESTMapper() meta.RESTMapper { return f.c.RESTMapper() }
func (f *alanYöneticiliİstemci) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return f.c.GroupVersionKindFor(obj)
}
func (f *alanYöneticiliİstemci) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return f.c.IsObjectNamespaced(obj)
}

func (f *alanYöneticiliİstemci) Status() StatusWriter {
	return &altKaynakYöneticiliİstemci{
		sahip:           f.sahip,
		altKaynakYazıcı: f.c.Status(),
	}
}

func (f *alanYöneticiliİstemci) SubResource(altKaynak string) SubResourceClient {
	c := f.c.SubResource(altKaynak)
	return &altKaynakYöneticiliİstemci{
		sahip:             f.sahip,
		altKaynakYazıcı:   c,
		SubResourceReader: c,
	}
}

type altKaynakYöneticiliİstemci struct {
	sahip           string
	altKaynakYazıcı SubResourceWriter
	SubResourceReader
}

func (f *altKaynakYöneticiliİstemci) Create(ctx context.Context, obj Object, altKaynak Object, opts ...SubResourceCreateOption) error {
	return f.altKaynakYazıcı.Create(ctx, obj, altKaynak, append([]SubResourceCreateOption{FieldOwner(f.sahip)}, opts...)...)
}

func (f *altKaynakYöneticiliİstemci) Update(ctx context.Context, obj Object, opts ...SubResourceUpdateOption) error {
	return f.altKaynakYazıcı.Update(ctx, obj, append([]SubResourceUpdateOption{FieldOwner(f.sahip)}, opts...)...)
}

func (f *altKaynakYöneticiliİstemci) Patch(ctx context.Context, obj Object, patch Patch, opts ...SubResourcePatchOption) error {
	return f.altKaynakYazıcı.Patch(ctx, obj, patch, append([]SubResourcePatchOption{FieldOwner(f.sahip)}, opts...)...)
}
