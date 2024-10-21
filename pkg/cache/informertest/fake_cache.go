/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMADAN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisansa bakınız.
*/

package informertest

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	toolscache "k8s.io/client-go/tools/cache"

	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllertest"
)

var _ cache.Cache = &SahteBilgilendiriciler{}

// SahteBilgilendiriciler, Bilgilendiricilerin sahte bir uygulamasıdır.
type SahteBilgilendiriciler struct {
	BilgilendiricilerGVK map[schema.GroupVersionKind]toolscache.SharedIndexInformer
	Sema                 *runtime.Scheme
	Hata                 error
	SenkronizeEdildi     *bool
}

// GetInformerForKind, Bilgilendiricileri uygular.
func (c *SahteBilgilendiriciler) GetInformerForKind(ctx context.Context, gvk schema.GroupVersionKind, opts ...cache.InformerGetOption) (cache.Informer, error) {
	if c.Sema == nil {
		c.Sema = scheme.Scheme
	}
	obj, err := c.Sema.New(gvk)
	if err != nil {
		return nil, err
	}
	return c.bilgilendiriciIcin(gvk, obj)
}

// SahteInformerForKind, Bilgilendiricileri uygular.
func (c *SahteBilgilendiriciler) SahteInformerForKind(ctx context.Context, gvk schema.GroupVersionKind) (*controllertest.FakeInformer, error) {
	i, err := c.GetInformerForKind(ctx, gvk)
	if err != nil {
		return nil, err
	}
	return i.(*controllertest.FakeInformer), nil
}

// GetInformer, Bilgilendiricileri uygular.
func (c *SahteBilgilendiriciler) GetInformer(ctx context.Context, obj client.Object, opts ...cache.InformerGetOption) (cache.Informer, error) {
	if c.Sema == nil {
		c.Sema = scheme.Scheme
	}
	gvks, _, err := c.Sema.ObjectKinds(obj)
	if err != nil {
		return nil, err
	}
	gvk := gvks[0]
	return c.bilgilendiriciIcin(gvk, obj)
}

// RemoveInformer, Bilgilendiricileri uygular.
func (c *SahteBilgilendiriciler) RemoveInformer(ctx context.Context, obj client.Object) error {
	if c.Sema == nil {
		c.Sema = scheme.Scheme
	}
	gvks, _, err := c.Sema.ObjectKinds(obj)
	if err != nil {
		return err
	}
	gvk := gvks[0]
	delete(c.BilgilendiricilerGVK, gvk)
	return nil
}

// WaitForCacheSync, Bilgilendiricileri uygular.
func (c *SahteBilgilendiriciler) WaitForCacheSync(ctx context.Context) bool {
	if c.SenkronizeEdildi == nil {
		return true
	}
	return *c.SenkronizeEdildi
}

// SahteInformerFor, Bilgilendiricileri uygular.
func (c *SahteBilgilendiriciler) SahteInformerFor(ctx context.Context, obj client.Object) (*controllertest.FakeInformer, error) {
	i, err := c.GetInformer(ctx, obj)
	if err != nil {
		return nil, err
	}
	return i.(*controllertest.FakeInformer), nil
}

func (c *SahteBilgilendiriciler) bilgilendiriciIcin(gvk schema.GroupVersionKind, _ runtime.Object) (toolscache.SharedIndexInformer, error) {
	if c.Hata != nil {
		return nil, c.Hata
	}
	if c.BilgilendiricilerGVK == nil {
		c.BilgilendiricilerGVK = map[schema.GroupVersionKind]toolscache.SharedIndexInformer{}
	}
	bilgilendirici, ok := c.BilgilendiricilerGVK[gvk]
	if ok {
		return bilgilendirici, nil
	}

	c.BilgilendiricilerGVK[gvk] = &controllertest.FakeInformer{}
	return c.BilgilendiricilerGVK[gvk], nil
}

// Start, Bilgilendiricileri uygular.
func (c *SahteBilgilendiriciler) Start(ctx context.Context) error {
	return c.Hata
}

// IndexField, Cache'i uygular.
func (c *SahteBilgilendiriciler) IndexField(ctx context.Context, obj client.Object, field string, extractValue client.IndexerFunc) error {
	return nil
}

// Get, Cache'i uygular.
func (c *SahteBilgilendiriciler) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return nil
}

// List, Cache'i uygular.
func (c *SahteBilgilendiriciler) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return nil
}
