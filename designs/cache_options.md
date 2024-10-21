Cache Options
===================

Bu belge, gelecekte önbellek seçeneklerinin nasıl görüneceğini hayal ettiğimiz şekilde açıklamaktadır.

## Amaçlar

* Herkesi desteklemek istediğimiz önbellek ayarları ve yapılandırma yüzeyi konusunda hizalamak
* Hem karmaşık önbellek kurulumlarını desteklediğimizden hem de sezgisel bir yapılandırma kullanıcı deneyimi sağladığımızdan emin olmak

## Amaç Dışı

* Önbelleğin tasarımını ve uygulanmasını açıklamak.
  Varsayım, en ayrıntılı seviyede "farklı seçicilere sahip çoklu ad alanları başına nesne" ile sonuçlanacağımız ve bunun mevcut çoklu ad alanı önbelleğini genişleterek bir "meta önbellek" kullanılarak uygulanabileceğidir.
* Bu ayarların ne zaman uygulanacağına dair herhangi bir zaman çizelgesi belirlemek.
  Uygulama, birisi gerçek işi yapmaya başladığında zamanla kademeli olarak gerçekleşecektir.

## Öneri

```go
const (
   AllNamespaces = corev1.NamespaceAll
)

type Config struct {
  LabelSelector labels.Selector
  FieldSelector fields.Selector
  Transform     toolscache.TransformFunc
  UnsafeDisableDeepCopy *bool
}

type ByObject struct {
  Namespaces map[string]*Config
  Config *Config
}

type Options struct {
  ByObject map[client.Object]*ByObject
  DefaultNamespaces map[string]*Config
  DefaultLabelSelector labels.Selector
  DefaultFieldSelector fields.Selector
  DefaultUnsafeDisableDeepCopy *bool
  DefaultTransform toolscache.TransformFunc
  HTTPClient *http.Client
  Scheme *runtime.Scheme
  Mapper meta.RESTMapper
  SyncPeriod *time.Duration
}
```

## Örnek Kullanımlar

### `public` ve `kube-system` ad alanlarındaki ConfigMap'leri ve `operator` ad alanındaki Secrets'leri önbelleğe almak

```go
cache.Options{
  ByObject: map[client.Object]*cache.ByObject{
    &corev1.ConfigMap{}: {
      Namespaces: map[string]*cache.Config{
        "public":      {},
        "kube-system": {},
      },
    },
    &corev1.Secret{}: {Namespaces: map[string]*Config{
        "operator": {},
    }},
  },
}
```

### Tüm ad alanlarındaki ConfigMap'leri seçicisiz önbelleğe almak, ancak `operator` ad alanı için bir seçiciye sahip olmak

```go
cache.Options{
  ByObject: map[client.Object]*cache.ByObject{
    &corev1.ConfigMap{}: {
      Namespaces: map[string]*cache.Config{
        cache.AllNamespaces: nil,
        "operator": {LabelSelector: labelSelector},
      },
    },
  },
}
```

### Ad alanlı nesneler için yalnızca `operator` ad alanını ve Dağıtımlar için tüm ad alanlarını önbelleğe almak

```go
cache.Options{
  ByObject: map[client.Object]*cache.ByObject{
    &appsv1.Deployment: {Namespaces: map[string]*cache.Config{
       cache.AllNamespaces: nil,
    }},
  },
  DefaultNamespaces: map[string]*cache.Config{
      "operator": nil,
  },
}
```

### Her şey için bir LabelSelector kullanmak, ancak Nodes için kullanmamak

```go
cache.Options{
  ByObject: map[client.Object]*cache.ByObject{
    &corev1.Node: {LabelSelector: labels.Everything()},
  },
  DefaultLabelSelector: myLabelSelector,
}
```

### Ad alanlı nesneleri yalnızca `foo` ve `bar` ad alanlarında önbelleğe almak

```go
cache.Options{
  DefaultNamespaces: map[string]*cache.Config{
    "foo": nil,
    "bar": nil,
  }
}
```
