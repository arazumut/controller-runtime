/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izinle gerekli olmadıkça,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ OLMAKSIZIN, açık veya zımni.
Lisans kapsamındaki izinleri ve sınırlamaları belirten
Lisans'a bakınız.
*/

package internal

import (
	"context"
	"fmt"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/internal/field/selector"
)

// CacheReader, client.Reader arayüzünü uygulayan bir yapıdır.
var _ client.Reader = &CacheReader{}

// CacheReader, tek bir tür için client.Reader arayüzünü uygulamak üzere cache.Index'i saran bir yapıdır.
type CacheReader struct {
	// indexer, bu önbellek tarafından sarılan temel indexer'dir.
	indexer cache.Indexer

	// groupVersionKind, kaynağın grup-sürüm-tür bilgisidir.
	groupVersionKind schema.GroupVersionKind

	// scopeName, kaynağın kapsamını belirtir (ad alanlı veya küme kapsamlı).
	scopeName apimeta.RESTScopeName

	// disableDeepCopy, nesneleri alırken veya listeleme sırasında derin kopyalama yapılmamasını belirtir.
	// Bu etkinleştirildiğinde, nesneyi değiştirmeden önce DeepCopy yapmanız gerekir,
	// aksi takdirde önbellekteki nesneyi değiştirmiş olursunuz.
	disableDeepCopy bool
}

// Get, indexer'da nesneyi kontrol eder ve bulunursa bir kopyasını yazar.
func (c *CacheReader) Get(_ context.Context, key client.ObjectKey, out client.Object, _ ...client.GetOption) error {
	if c.scopeName == apimeta.RESTScopeNameRoot {
		key.Namespace = ""
	}
	storeKey := objectKeyToStoreKey(key)

	// Nesneyi indexer önbelleğinden arayın
	obj, exists, err := c.indexer.GetByKey(storeKey)
	if err != nil {
		return err
	}

	// Bulunamadı, hata döndür
	if !exists {
		return apierrors.NewNotFound(schema.GroupResource{
			Group: c.groupVersionKind.Group,
			// Hata mesajında Tür olarak ayarlandığı için bu sorun değil
			Resource: c.groupVersionKind.Kind,
		}, key.Name)
	}

	// Sonucun bir runtime.Object olduğunu doğrulayın
	if _, isObj := obj.(runtime.Object); !isObj {
		// Bu asla olmamalı
		return fmt.Errorf("önbellek %T içeriyordu, bu bir Nesne değil", obj)
	}

	if c.disableDeepCopy {
		// derin kopyalamayı atla, bu güvensiz olabilir
		// nesneyi dışarıda değiştirmeden önce DeepCopy yapmanız gerekir
	} else {
		// önbelleği değiştirmemek için derin kopyalama yap
		obj = obj.(runtime.Object).DeepCopyObject()
	}

	// Önbellekteki öğenin değerini döndürülen değere kopyalayın
	// TODO(directxman12): bu korkunç bir hack, lütfen düzeltin (deepcopyinto yapmalıyız)
	outVal := reflect.ValueOf(out)
	objVal := reflect.ValueOf(obj)
	if !objVal.Type().AssignableTo(outVal.Type()) {
		return fmt.Errorf("önbellekte %s türü vardı, ancak %s istendi", objVal.Type(), outVal.Type())
	}
	reflect.Indirect(outVal).Set(reflect.Indirect(objVal))
	if !c.disableDeepCopy {
		out.GetObjectKind().SetGroupVersionKind(c.groupVersionKind)
	}

	return nil
}

// List, indexer'dan öğeleri listeler ve bunları out'a yazar.
func (c *CacheReader) List(_ context.Context, out client.ObjectList, opts ...client.ListOption) error {
	var objs []interface{}
	var err error

	listOpts := client.ListOptions{}
	listOpts.ApplyOptions(opts)

	if listOpts.Continue != "" {
		return fmt.Errorf("continue list seçeneği önbellek tarafından desteklenmiyor")
	}

	switch {
	case listOpts.FieldSelector != nil:
		requiresExact := selector.RequiresExactMatch(listOpts.FieldSelector)
		if !requiresExact {
			return fmt.Errorf("kesin olmayan alan eşleşmeleri önbellek tarafından desteklenmiyor")
		}
		// alan seçiciye göre tüm nesneleri listeleyin. Bu ad alanlıysa ve bir tane varsa, ad alanlı indeks anahtarını isteyin.
		// Aksi takdirde, sahte "tüm ad alanları" ad alanını kullanarak ad alanlı olmayan varyantı isteyin.
		objs, err = byIndexes(c.indexer, listOpts.FieldSelector.Requirements(), listOpts.Namespace)
	case listOpts.Namespace != "":
		objs, err = c.indexer.ByIndex(cache.NamespaceIndex, listOpts.Namespace)
	default:
		objs = c.indexer.List()
	}
	if err != nil {
		return err
	}
	var labelSel labels.Selector
	if listOpts.LabelSelector != nil {
		labelSel = listOpts.LabelSelector
	}

	limitSet := listOpts.Limit > 0

	runtimeObjs := make([]runtime.Object, 0, len(objs))
	for _, item := range objs {
		// Limit seçeneği ayarlandıysa ve listelenen öğe sayısı bu limiti aşıyorsa, okumayı durdurun.
		if limitSet && int64(len(runtimeObjs)) >= listOpts.Limit {
			break
		}
		obj, isObj := item.(runtime.Object)
		if !isObj {
			return fmt.Errorf("önbellek %T içeriyordu, bu bir Nesne değil", item)
		}
		meta, err := apimeta.Accessor(obj)
		if err != nil {
			return err
		}
		if labelSel != nil {
			lbls := labels.Set(meta.GetLabels())
			if !labelSel.Matches(lbls) {
				continue
			}
		}

		var outObj runtime.Object
		if c.disableDeepCopy || (listOpts.UnsafeDisableDeepCopy != nil && *listOpts.UnsafeDisableDeepCopy) {
			// derin kopyalamayı atla, bu güvensiz olabilir
			// nesneyi dışarıda değiştirmeden önce DeepCopy yapmanız gerekir
			outObj = obj
		} else {
			outObj = obj.DeepCopyObject()
			outObj.GetObjectKind().SetGroupVersionKind(c.groupVersionKind)
		}
		runtimeObjs = append(runtimeObjs, outObj)
	}
	return apimeta.SetList(out, runtimeObjs)
}

func byIndexes(indexer cache.Indexer, requires fields.Requirements, namespace string) ([]interface{}, error) {
	var (
		err  error
		objs []interface{}
		vals []string
	)
	indexers := indexer.GetIndexers()
	for idx, req := range requires {
		indexName := FieldIndexName(req.Field)
		indexedValue := KeyToNamespacedKey(namespace, req.Value)
		if idx == 0 {
			// ilk gereksinimi kullanarak anlık veri alıyoruz
			// TODO(halfcrazy): client-go karmaşık indeks sağladığında karmaşık indeksi kullan
			// https://github.com/kubernetes/kubernetes/issues/109329
			objs, err = indexer.ByIndex(indexName, indexedValue)
			if err != nil {
				return nil, err
			}
			if len(objs) == 0 {
				return nil, nil
			}
			continue
		}
		fn, exist := indexers[indexName]
		if !exist {
			return nil, fmt.Errorf("%s adında bir indeks yok", indexName)
		}
		filteredObjects := make([]interface{}, 0, len(objs))
		for _, obj := range objs {
			vals, err = fn(obj)
			if err != nil {
				return nil, err
			}
			for _, val := range vals {
				if val == indexedValue {
					filteredObjects = append(filteredObjects, obj)
					break
				}
			}
		}
		if len(filteredObjects) == 0 {
			return nil, nil
		}
		objs = filteredObjects
	}
	return objs, nil
}

// objectKeyToStorageKey, bir nesne anahtarını depolama anahtarına dönüştürür.
// MetaNamespaceKeyFunc'a benzer. Bu, anahtar formatını MetaNamespaceKeyFunc ile kolayca senkronize tutmak için ayrıdır.
func objectKeyToStoreKey(k client.ObjectKey) string {
	if k.Namespace == "" {
		return k.Name
	}
	return k.Namespace + "/" + k.Name
}

// FieldIndexName, bir indexer ile kullanım için verilen alan üzerindeki indeksin adını oluşturur.
func FieldIndexName(field string) string {
	return "field:" + field
}

// allNamespacesNamespace, tüm ad alanları arasında listelemek istediğimizde "ad alanı" olarak kullanılır.
const allNamespacesNamespace = "__all_namespaces"

// KeyToNamespacedKey, alan seçici indekslerinde kullanım için verilen indeks anahtarını bir ad alanı ile öne ekler.
func KeyToNamespacedKey(ns string, baseKey string) string {
	if ns != "" {
		return ns + "/" + baseKey
	}
	return allNamespacesNamespace + "/" + baseKey
}
