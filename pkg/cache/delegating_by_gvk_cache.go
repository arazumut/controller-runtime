/*
2023 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği olmadıkça,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
yetkiler ve sınırlamalar için Lisansa bakınız.
*/

package cache

import (
	"context"
	"strings"
	"sync"

	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// delegatingByGVKCache, belirli bir tür önbelleğine devredilir
// ve aksi takdirde defaultCache kullanır.
type delegatingByGVKCache struct {
	scheme       *runtime.Scheme
	caches       map[schema.GroupVersionKind]Cache
	defaultCache Cache
}

func (dbt *delegatingByGVKCache) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	cache, err := dbt.cacheForObject(obj)
	if err != nil {
		return err
	}
	return cache.Get(ctx, key, obj, opts...)
}

func (dbt *delegatingByGVKCache) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	cache, err := dbt.cacheForObject(list)
	if err != nil {
		return err
	}
	return cache.List(ctx, list, opts...)
}

func (dbt *delegatingByGVKCache) RemoveInformer(ctx context.Context, obj client.Object) error {
	cache, err := dbt.cacheForObject(obj)
	if err != nil {
		return err
	}
	return cache.RemoveInformer(ctx, obj)
}

func (dbt *delegatingByGVKCache) GetInformer(ctx context.Context, obj client.Object, opts ...InformerGetOption) (Informer, error) {
	cache, err := dbt.cacheForObject(obj)
	if err != nil {
		return nil, err
	}
	return cache.GetInformer(ctx, obj, opts...)
}

func (dbt *delegatingByGVKCache) GetInformerForKind(ctx context.Context, gvk schema.GroupVersionKind, opts ...InformerGetOption) (Informer, error) {
	return dbt.cacheForGVK(gvk).GetInformerForKind(ctx, gvk, opts...)
}

func (dbt *delegatingByGVKCache) Start(ctx context.Context) error {
	tümCaches := maps.Values(dbt.caches)
	tümCaches = append(tümCaches, dbt.defaultCache)

	wg := &sync.WaitGroup{}
	errs := make(chan error)
	for idx := range tümCaches {
		cache := tümCaches[idx]
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := cache.Start(ctx); err != nil {
				errs <- err
			}
		}()
	}

	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
		wg.Wait()
		return nil
	}
}

func (dbt *delegatingByGVKCache) WaitForCacheSync(ctx context.Context) bool {
	senkronize := true
	for _, cache := range append(maps.Values(dbt.caches), dbt.defaultCache) {
		if !cache.WaitForCacheSync(ctx) {
			senkronize = false
		}
	}

	return senkronize
}

func (dbt *delegatingByGVKCache) IndexField(ctx context.Context, obj client.Object, field string, extractValue client.IndexerFunc) error {
	cache, err := dbt.cacheForObject(obj)
	if err != nil {
		return err
	}
	return cache.IndexField(ctx, obj, field, extractValue)
}

func (dbt *delegatingByGVKCache) cacheForObject(o runtime.Object) (Cache, error) {
	gvk, err := apiutil.GVKForObject(o, dbt.scheme)
	if err != nil {
		return nil, err
	}
	gvk.Kind = strings.TrimSuffix(gvk.Kind, "List")
	return dbt.cacheForGVK(gvk), nil
}

func (dbt *delegatingByGVKCache) cacheForGVK(gvk schema.GroupVersionKind) Cache {
	if specific, hasSpecific := dbt.caches[gvk]; hasSpecific {
		return specific
	}

	return dbt.defaultCache
}
