# Filter cache ListWatch using selectors

## Motivasyon

Controller-Runtime denetleyicileri, Kubernetes nesnelerinden gelen olaylara abone olmak ve bu nesneleri daha verimli bir şekilde okumak için bir önbellek kullanır, böylece API'ye çağrı yapmaktan kaçınır. Bu önbellek, Kubernetes informers tarafından desteklenir.

Bu önbelleği filtrelemenin tek yolu, ad alanı ve kaynak türü ile sınırlıdır. Bir denetleyicinin yalnızca küçük bir nesne alt kümesiyle ilgilendiği durumlarda (örneğin, bir düğümdeki tüm podlar), bu yeterince verimli olmayabilir.

Filtreyi karşılamayan nesneler için filtrelenmiş bir önbellek tarafından desteklenen bir istemciye yapılan istekler hiçbir şey döndürmez, bu yüzden kullanıcıları yalnızca ne yaptıklarını bildiklerinden emin olduklarında bunu kullanmaları konusunda uygun şekilde uyarmamız gerekir.

Bu öneri, "Kullanıcılar, önbellek filtrelerini karşılamayan bir şey istediklerinde geri bildirim alacak şekilde önbellek destekli istemciye bunu nasıl ekleriz" sorununu, yalnızca önbellek paketinde filtre mantığını uygulayarak atlar. Bu, ileri düzey kullanıcıların filtrelenmiş bir önbelleği yönetici ve küme paketindeki mevcut `NewCacheFunc` seçeneği ile birleştirmelerine olanak tanır ve aynı zamanda, sonuçları ve ilgili riskleri farkında olmayan yeni kullanıcılardan gizler.

Bugün, controller-runtime ile filtrelenmiş bir önbellek elde etmenin tek alternatifi, bunu dışarıda inşa etmektir. Çünkü böyle bir önbellek, mevcut önbelleği çoğunlukla kopyalar ve sadece bazı seçenekler ekler, bu tüketiciler için pek iyi değildir.

Bu öneri, şu konuyla ilgilidir [2]

## Öneri

Ortak yapılar ve yardımcılarla `pkg/cache/internal/selector.go`'da yeni bir seçici kodu ekleyin

```golang
package internal

...

// SelectorsByObject bir runtime.Object'i bir alan/etiket seçici ile ilişkilendirir
type SelectorsByObject map[client.Object]Selector

// SelectorsByGVK bir GroupVersionResource'u bir alan/etiket seçici ile ilişkilendirir
type SelectorsByGVK map[schema.GroupVersionKind]Selector

// Selector, ListOptions'ı doldurmak için etiket/alan seçicisini belirtir
type Selector struct {
    Label labels.Selector
    Field fields.Selector
}

// ApplyToList, gerekirse ListOptions LabelSelector ve FieldSelector'ı doldurur
func (s Selector) ApplyToList(listOpts *metav1.ListOptions) {
...
}
```

`pkg/cache/cache.go`'ya içsel bir tür takma adı ekleyin

```golang
type SelectorsByObject internal.SelectorsByObject
```

`cache.Options`'ı şu şekilde genişletin:

```golang
type Options struct {
    Scheme            *runtime.Scheme
    Mapper            meta.RESTMapper
    Resync            *time.Duration
    Namespace         string
    SelectorsByObject SelectorsByObject
}
```

Yeni bir oluşturucu fonksiyonu ekleyin, bu fonksiyon cache.Options kullanarak bir önbellek oluşturucu döndürecek, kullanıcılar burada SelectorsByObject ayarlayarak önbelleği filtreleyebilir, bu SelectorByObject'i SelectorsByGVK'ya dönüştürecektir

```golang
func BuilderWithOptions(options cache.Options) NewCacheFunc {
...
}
```

informer's ListWatch'a geçirildi ve filtreleme seçeneği eklendi:

```golang

# At pkg/cache/internal/informers_map.go

ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
            ip.selectors[gvk].ApplyToList(&opts)
...
```

İşte `pkg/cache` kısmındaki implementasyon ile ilgili bir PR [3]

## Örnek

Kullanıcı, varsayılan olandan farklı bir önbellek kullanmanın sonuçlarını tam olarak bildiklerini açıkça belirtmek için `NewCache` fonksiyonunu geçersiz kılacaktır

```golang
 ctrl.Options.NewCache = cache.BuilderWithOptions(cache.Options{
                            SelectorsByObject: cache.SelectorsByObject{
                                    &corev1.Node{}: {
                                        Field: fields.SelectorFromSet(fields.Set{"metadata.name": "node01"}),
                                    }
                                    &v1beta1.NodeNetworkState{}: {
                                        Field: fields.SelectorFromSet(fields.Set{"metadata.name": "node01"}),
                                        Label: labels.SelectorFromSet(labels.Set{"app": "kubernetes-nmstate"}),
                                    }
                                }
                            }
                        )
```

[1] https://github.com/nmstate/kubernetes-nmstate/pull/687
[2] https://github.com/kubernetes-sigs/controller-runtime/issues/244
[3] https://github.com/kubernetes-sigs/controller-runtime/pull/1404

