/*
2018 Kubernetes Yazarları tarafından yazılmıştır.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa uyarınca veya yazılı olarak kabul edilmediği sürece,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMADAN.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisansa bakınız.
*/

package cache

import (
	"context"
	"fmt"
	"strings"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"

	"sigs.k8s.io/controller-runtime/pkg/cache/internal"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

var (
	_ Informers     = &informerCache{}
	_ client.Reader = &informerCache{}
	_ Cache         = &informerCache{}
)

// ErrCacheNotStarted, önbellekten okuma yapmaya çalışırken önbelleğin başlatılmadığını belirten hatadır.
type ErrCacheNotStarted struct{}

func (*ErrCacheNotStarted) Error() string {
	return "önbellek başlatılmadı, nesneler okunamıyor"
}

var _ error = (*ErrCacheNotStarted)(nil)

// ErrResourceNotCached, istemcinin önbellekten istediği kaynak türünün önbelleğe alınmadığını belirtir.
type ErrResourceNotCached struct {
	GVK schema.GroupVersionKind
}

// Error, hatayı döndürür.
func (r ErrResourceNotCached) Error() string {
	return fmt.Sprintf("%s önbelleğe alınmadı", r.GVK.String())
}

var _ error = (*ErrResourceNotCached)(nil)

// informerCache, internal.Informers'dan doldurulan bir Kubernetes Nesne önbelleğidir.
// informerCache, internal.Informers'ı sarmalar.
type informerCache struct {
	scheme *runtime.Scheme
	*internal.Informers
	readerFailOnMissingInformer bool
}

// Get, Reader'ı uygular.
func (ic *informerCache) Get(ctx context.Context, key client.ObjectKey, out client.Object, opts ...client.GetOption) error {
	gvk, err := apiutil.GVKForObject(out, ic.scheme)
	if err != nil {
		return err
	}

	started, cache, err := ic.getInformerForKind(ctx, gvk, out)
	if err != nil {
		return err
	}

	if !started {
		return &ErrCacheNotStarted{}
	}
	return cache.Reader.Get(ctx, key, out, opts...)
}

// List, Reader'ı uygular.
func (ic *informerCache) List(ctx context.Context, out client.ObjectList, opts ...client.ListOption) error {
	gvk, cacheTypeObj, err := ic.objectTypeForListObject(out)
	if err != nil {
		return err
	}

	started, cache, err := ic.getInformerForKind(ctx, *gvk, cacheTypeObj)
	if err != nil {
		return err
	}

	if !started {
		return &ErrCacheNotStarted{}
	}

	return cache.Reader.List(ctx, out, opts...)
}

// objectTypeForListObject, verilen liste türüne karşılık gelen tek bir nesne için runtime.Object ve ilgili GVK'yı bulmaya çalışır.
func (ic *informerCache) objectTypeForListObject(list client.ObjectList) (*schema.GroupVersionKind, runtime.Object, error) {
	gvk, err := apiutil.GVKForObject(list, ic.scheme)
	if err != nil {
		return nil, nil, err
	}

	// Liste türünün sonundaki "List" kısmını çıkararak liste olmayan GVK'yı elde edin.
	gvk.Kind = strings.TrimSuffix(gvk.Kind, "List")

	// unstructured.UnstructuredList'i işleyin.
	if _, isUnstructured := list.(runtime.Unstructured); isUnstructured {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(gvk)
		return &gvk, u, nil
	}
	// metav1.PartialObjectMetadataList'i işleyin.
	if _, isPartialObjectMetadata := list.(*metav1.PartialObjectMetadataList); isPartialObjectMetadata {
		pom := &metav1.PartialObjectMetadata{}
		pom.SetGroupVersionKind(gvk)
		return &gvk, pom, nil
	}

	// Diğer liste türlerinin şemada kayıtlı karşılık gelen liste olmayan türleri olmalıdır.
	// Bu türden yeni bir örnek oluşturmak için bunu kullanın.
	cacheTypeObj, err := ic.scheme.New(gvk)
	if err != nil {
		return nil, nil, err
	}
	return &gvk, cacheTypeObj, nil
}

func applyGetOptions(opts ...InformerGetOption) *internal.GetOptions {
	cfg := &InformerGetOptions{}
	for _, opt := range opts {
		opt(cfg)
	}
	return (*internal.GetOptions)(cfg)
}

// GetInformerForKind, GroupVersionKind için bilgi vericiyi döndürür. Eğer bilgi verici yoksa, biri başlatılır.
func (ic *informerCache) GetInformerForKind(ctx context.Context, gvk schema.GroupVersionKind, opts ...InformerGetOption) (Informer, error) {
	// gvk'yı bir nesneye eşleyin
	obj, err := ic.scheme.New(gvk)
	if err != nil {
		return nil, err
	}

	_, i, err := ic.Informers.Get(ctx, gvk, obj, applyGetOptions(opts...))
	if err != nil {
		return nil, err
	}
	return i.Informer, nil
}

// GetInformer, obj için bilgi vericiyi döndürür. Eğer bilgi verici yoksa, biri başlatılır.
func (ic *informerCache) GetInformer(ctx context.Context, obj client.Object, opts ...InformerGetOption) (Informer, error) {
	gvk, err := apiutil.GVKForObject(obj, ic.scheme)
	if err != nil {
		return nil, err
	}

	_, i, err := ic.Informers.Get(ctx, gvk, obj, applyGetOptions(opts...))
	if err != nil {
		return nil, err
	}
	return i.Informer, nil
}

func (ic *informerCache) getInformerForKind(ctx context.Context, gvk schema.GroupVersionKind, obj runtime.Object) (bool, *internal.Cache, error) {
	if ic.readerFailOnMissingInformer {
		cache, started, ok := ic.Informers.Peek(gvk, obj)
		if !ok {
			return false, nil, &ErrResourceNotCached{GVK: gvk}
		}
		return started, cache, nil
	}

	return ic.Informers.Get(ctx, gvk, obj, &internal.GetOptions{})
}

// RemoveInformer, bilgi vericiyi devre dışı bırakır ve önbellekten kaldırır.
func (ic *informerCache) RemoveInformer(_ context.Context, obj client.Object) error {
	gvk, err := apiutil.GVKForObject(obj, ic.scheme)
	if err != nil {
		return err
	}

	ic.Informers.Remove(gvk, obj)
	return nil
}

// NeedLeaderElection, bu işlemin lider kilidi gerektirmeden başlatılabileceğini belirtmek için LeaderElectionRunnable arayüzünü uygular.
func (ic *informerCache) NeedLeaderElection() bool {
	return false
}

// IndexField, verilen alanın değerlerini almak için extractValue işlevini kullanarak temel bilgi vericiye bir indeks ekler.
// Bu indeks daha sonra List'e bir alan seçici geçirerek kullanılabilir. "Normal" alan seçicileriyle birebir uyumluluk için yalnızca bir değer döndürün.
// Değerler herhangi bir şey olabilir. Otomatik olarak nesnenin ad alanı ile öneklenecektir, eğer varsa.
// Geçilen nesnelerin doğru türde nesneler olduğu garanti edilir.
func (ic *informerCache) IndexField(ctx context.Context, obj client.Object, field string, extractValue client.IndexerFunc) error {
	informer, err := ic.GetInformer(ctx, obj)
	if err != nil {
		return err
	}
	return indexByField(informer, field, extractValue)
}

func indexByField(informer Informer, field string, extractValue client.IndexerFunc) error {
	indexFunc := func(objRaw interface{}) ([]string, error) {
		obj, isObj := objRaw.(client.Object)
		if !isObj {
			return nil, fmt.Errorf("nesne türü %T olan nesne bir Nesne değil", objRaw)
		}
		meta, err := apimeta.Accessor(obj)
		if err != nil {
			return nil, err
		}
		ns := meta.GetNamespace()

		rawVals := extractValue(obj)
		var vals []string
		if ns == "" {
			vals = make([]string, len(rawVals))
		} else {
			vals = make([]string, len(rawVals)*2)
		}
		for i, rawVal := range rawVals {
			vals[i] = internal.KeyToNamespacedKey(ns, rawVal)
			if ns != "" {
				vals[i+len(rawVals)] = internal.KeyToNamespacedKey("", rawVal)
			}
		}

		return vals, nil
	}

	return informer.AddIndexers(cache.Indexers{internal.FieldIndexName(field): indexFunc})
}
