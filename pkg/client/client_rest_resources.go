/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa uyarınca veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
yetkiler ve sınırlamalar için Lisansa bakınız.
*/

package client

import (
	"net/http"
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// clientRestResources Kubernetes türleri için rest istemcileri ve meta verileri oluşturur ve saklar.
type clientRestResources struct {
	httpClient                 *http.Client
	config                     *rest.Config
	scheme                     *runtime.Scheme
	mapper                     meta.RESTMapper
	codecs                     serializer.CodecFactory
	structuredResourceByType   map[schema.GroupVersionKind]*resourceMeta
	unstructuredResourceByType map[schema.GroupVersionKind]*resourceMeta
	mu                         sync.RWMutex
}

// newResource objeyi bir Kubernetes Kaynağına eşler ve bu Kaynak için bir istemci oluşturur.
// Eğer nesne bir liste ise, kaynak öğenin türünü temsil eder.
func (c *clientRestResources) newResource(gvk schema.GroupVersionKind, isList, isUnstructured bool) (*resourceMeta, error) {
	if strings.HasSuffix(gvk.Kind, "List") && isList {
		gvk.Kind = gvk.Kind[:len(gvk.Kind)-4]
	}

	client, err := apiutil.RESTClientForGVK(gvk, isUnstructured, c.config, c.codecs, c.httpClient)
	if err != nil {
		return nil, err
	}
	mapping, err := c.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	return &resourceMeta{Interface: client, mapping: mapping, gvk: gvk}, nil
}

// getResource verilen nesne türü için kaynak meta bilgilerini döndürür.
// Eğer nesne bir liste ise, kaynak öğenin türünü temsil eder.
func (c *clientRestResources) getResource(obj runtime.Object) (*resourceMeta, error) {
	gvk, err := apiutil.GVKForObject(obj, c.scheme)
	if err != nil {
		return nil, err
	}

	_, isUnstructured := obj.(runtime.Unstructured)

	c.mu.RLock()
	resourceByType := c.structuredResourceByType
	if isUnstructured {
		resourceByType = c.unstructuredResourceByType
	}
	r, known := resourceByType[gvk]
	c.mu.RUnlock()

	if known {
		return r, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	r, err = c.newResource(gvk, meta.IsListType(obj), isUnstructured)
	if err != nil {
		return nil, err
	}
	resourceByType[gvk] = r
	return r, err
}

// getObjMeta objMeta döndürür, hem tür hem de nesne meta verilerini ve durumunu içerir.
func (c *clientRestResources) getObjMeta(obj runtime.Object) (*objMeta, error) {
	r, err := c.getResource(obj)
	if err != nil {
		return nil, err
	}
	m, err := meta.Accessor(obj)
	if err != nil {
		return nil, err
	}
	return &objMeta{resourceMeta: r, Object: m}, err
}

// resourceMeta bir Kubernetes türü için durumu saklar.
type resourceMeta struct {
	rest.Interface
	gvk     schema.GroupVersionKind
	mapping *meta.RESTMapping
}

// isNamespaced türün ad alanlı olup olmadığını döndürür.
func (r *resourceMeta) isNamespaced() bool {
	return r.mapping.Scope.Name() != meta.RESTScopeNameRoot
}

// resource türün kaynak adını döndürür.
func (r *resourceMeta) resource() string {
	return r.mapping.Resource.Resource
}

// objMeta bir Kubernetes türü hakkında tür ve nesne bilgilerini saklar.
type objMeta struct {
	*resourceMeta
	metav1.Object
}
