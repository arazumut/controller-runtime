/*
2017 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ OLMAKSIZIN, açık veya zımni.
Lisans kapsamındaki izinler ve
sınırlamalar hakkında daha fazla bilgi için Lisansa bakın.
*/

package controllertest

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

var _ cache.SharedIndexInformer = &FakeInformer{}

// FakeInformer, testler için sahte Informer işlevselliği sağlar.
type FakeInformer struct {
	// Synced, Informer arayüzünü uygulamak için HasSynced işlevleri tarafından döndürülür
	Synced bool

	// RunCount, RunInformersAndControllers her çağrıldığında artırılır
	RunCount int

	handlers []eventHandlerWrapper
}

type modernResourceEventHandler interface {
	OnAdd(obj interface{}, isInInitialList bool)
	OnUpdate(oldObj, newObj interface{})
	OnDelete(obj interface{})
}

type legacyResourceEventHandler interface {
	OnAdd(obj interface{})
	OnUpdate(oldObj, newObj interface{})
	OnDelete(obj interface{})
}

// eventHandlerWrapper, client-go 1.27+ ve daha eski sürümlerle uyumlu bir şekilde ResourceEventHandler'ı sarar.
// Bu sürümlerde arayüz değiştirildi.
type eventHandlerWrapper struct {
	handler any
}

func (e eventHandlerWrapper) OnAdd(obj interface{}) {
	if m, ok := e.handler.(modernResourceEventHandler); ok {
		m.OnAdd(obj, false)
		return
	}
	e.handler.(legacyResourceEventHandler).OnAdd(obj)
}

func (e eventHandlerWrapper) OnUpdate(oldObj, newObj interface{}) {
	if m, ok := e.handler.(modernResourceEventHandler); ok {
		m.OnUpdate(oldObj, newObj)
		return
	}
	e.handler.(legacyResourceEventHandler).OnUpdate(oldObj, newObj)
}

func (e eventHandlerWrapper) OnDelete(obj interface{}) {
	if m, ok := e.handler.(modernResourceEventHandler); ok {
		m.OnDelete(obj)
		return
	}
	e.handler.(legacyResourceEventHandler).OnDelete(obj)
}

// AddIndexers hiçbir şey yapmaz.  TODO(community): Bunu uygulayın.
func (f *FakeInformer) AddIndexers(indexers cache.Indexers) error {
	return nil
}

// GetIndexer hiçbir şey yapmaz.  TODO(community): Bunu uygulayın.
func (f *FakeInformer) GetIndexer() cache.Indexer {
	return nil
}

// Informer, sahte Informer'ı döndürür.
func (f *FakeInformer) Informer() cache.SharedIndexInformer {
	return f
}

// HasSynced, Informer arayüzünü uygular.  f.Synced'i döndürür.
func (f *FakeInformer) HasSynced() bool {
	return f.Synced
}

// AddEventHandler, Informer arayüzünü uygular.  Sahte Informer'lara bir EventHandler ekler. TODO(community): Kayıt işlemini uygulayın.
func (f *FakeInformer) AddEventHandler(handler cache.ResourceEventHandler) (cache.ResourceEventHandlerRegistration, error) {
	f.handlers = append(f.handlers, eventHandlerWrapper{handler})
	return nil, nil
}

// Run, Informer arayüzünü uygular.  f.RunCount'u artırır.
func (f *FakeInformer) Run(<-chan struct{}) {
	f.RunCount++
}

// Add, obj için sahte bir Ekleme olayı oluşturur.
func (f *FakeInformer) Add(obj metav1.Object) {
	for _, h := range f.handlers {
		h.OnAdd(obj)
	}
}

// Update, obj için sahte bir Güncelleme olayı oluşturur.
func (f *FakeInformer) Update(oldObj, newObj metav1.Object) {
	for _, h := range f.handlers {
		h.OnUpdate(oldObj, newObj)
	}
}

// Delete, obj için sahte bir Silme olayı oluşturur.
func (f *FakeInformer) Delete(obj metav1.Object) {
	for _, h := range f.handlers {
		h.OnDelete(obj)
	}
}

// AddEventHandlerWithResyncPeriod hiçbir şey yapmaz.  TODO(community): Bunu uygulayın.
func (f *FakeInformer) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, resyncPeriod time.Duration) (cache.ResourceEventHandlerRegistration, error) {
	return nil, nil
}

// RemoveEventHandler hiçbir şey yapmaz.  TODO(community): Bunu uygulayın.
func (f *FakeInformer) RemoveEventHandler(handle cache.ResourceEventHandlerRegistration) error {
	return nil
}

// GetStore hiçbir şey yapmaz.  TODO(community): Bunu uygulayın.
func (f *FakeInformer) GetStore() cache.Store {
	return nil
}

// GetController hiçbir şey yapmaz.  TODO(community): Bunu uygulayın.
func (f *FakeInformer) GetController() cache.Controller {
	return nil
}

// LastSyncResourceVersion hiçbir şey yapmaz.  TODO(community): Bunu uygulayın.
func (f *FakeInformer) LastSyncResourceVersion() string {
	return ""
}

// SetWatchErrorHandler hiçbir şey yapmaz.  TODO(community): Bunu uygulayın.
func (f *FakeInformer) SetWatchErrorHandler(cache.WatchErrorHandler) error {
	return nil
}

// SetTransform hiçbir şey yapmaz.  TODO(community): Bunu uygulayın.
func (f *FakeInformer) SetTransform(t cache.TransformFunc) error {
	return nil
}

// IsStopped hiçbir şey yapmaz.  TODO(community): Bunu uygulayın.
func (f *FakeInformer) IsStopped() bool {
	return false
}
