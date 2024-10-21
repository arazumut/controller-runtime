# ComponentConfig Controller Runtime Support
Yazar: @christopherhein

Son Güncelleme: 03/02/2020

## İçindekiler

<!--ts-->
	* [ComponentConfig Controller Runtime Desteği](#componentconfig-controller-runtime-desteği)
		* [İçindekiler](#içindekiler)
		* [Özet](#özet)
		* [Motivasyon](#motivasyon)
			* [Açık Sorunlara Bağlantılar](#açık-sorunlara-bağlantılar)
			* [Hedefler](#hedefler)
			* [Hedef Dışı/Gelecek Çalışmalar](#hedef-dışıgelecek-çalışmalar)
		* [Teklif](#teklif)
		* [ComponentConfig Yükleme Sırası](#componentconfig-yükleme-sırası)
		* [Gömülebilir ComponentConfig Türü](#gömülebilir-componentconfig-türü)
		* [Varsayılan ComponentConfig Türü](#varsayılan-componentconfig-türü)
		* [ComponentConfig ile Bayrak Kullanımı](#componentconfig-ile-bayrak-kullanımı)
		* [Kubebuilder İskelet Örneği](#kubebuilder-iskelet-örneği)
		* [Kullanıcı Hikayeleri](#kullanıcı-hikayeleri)
			* [Varsayılan tür ile controller-runtime kullanan Controller Yazarı](#varsayılan-tür-ile-controller-runtime-kullanan-controller-yazarı)
			* [Özel tür ile controller-runtime kullanan Controller Yazarı](#özel-tür-ile-controller-runtime-kullanan-controller-yazarı)
			* [kubebuilder ile Controller Yazarı (kubebuilder için tbd teklif)](#kubebuilder-ile-controller-yazarı-kubebuilder-için-tbd-teklif)
			* [Yapılandırma değişiklikleri olmadan Controller Kullanıcısı](#yapılandırma-değişiklikleri-olmadan-controller-kullanıcısı)
			* [Yapılandırma değişiklikleri ile Controller Kullanıcısı](#yapılandırma-değişiklikleri-ile-controller-kullanıcısı)
		* [Riskler ve Azaltmalar](#riskler-ve-azaltmalar)
		* [Alternatifler](#alternatifler)
		* [Uygulama Geçmişi](#uygulama-geçmişi)

<!--te-->

## Özet

Şu anda `controller-runtime` kullanan kontroller, `ctrl.Manager`'ı yapılandırmak için bayraklar kullanmalı veya değerleri başlatma yöntemlerine sabitlemelidir. Çekirdek Kubernetes, bileşenleri yapılandırma mekanizması olarak bayrakları kullanmaktan uzaklaşmaya ve [`ComponentConfig` veya Sürüm Kontrollü Bileşen Yapılandırma Dosyaları](https://docs.google.com/document/d/1FdaEJUEh091qf5B98HM6_8MS764iXrxxigNIdwHYW9c/edit) üzerinde standartlaşmaya başladı. Bu teklif, bayrakları azaltarak ve CLI argümanlarını değiştirmelerini gerektirmeden kod tabanlı araçların kontrolleri kolayca yapılandırmasına izin vererek `controller-runtime`'a `ComponentConfig` getirmeyi amaçlamaktadır.

## Motivasyon

Bu değişiklik önemlidir çünkü:
- kontrollerin diğer makine süreçleri tarafından yapılandırılmasını kolaylaştıracaktır
- bir kontrolörü başlatmak için gereken bayrakları azaltacaktır
- bayraklar tarafından doğal olarak desteklenmeyen yapılandırma türlerine izin verecektir
- bayraklarda kırılma değişikliklerinden kaçınarak eski yapılandırmaları kullanma ve yükseltme imkanı tanıyacaktır

### Açık Sorunlara Bağlantılar

- [#518 Manager'ı ayarlamak için bir ComponentConfig sağlayın](https://github.com/kubernetes-sigs/controller-runtime/issues/518)
- [#207 Komut satırı bayrak şablonunu azaltın](https://github.com/kubernetes-sigs/controller-runtime/issues/207)
- [#722 Varsayılan olarak ComponentConfig'i uygulayın ve (çoğu) bayrakları kullanmayı bırakın](https://github.com/kubernetes-sigs/kubebuilder/issues/722)

### Hedefler

- Açık `ComponentConfig` türlerinden yapılandırma verilerini çekmek için bir arayüz sağlamak (aşağıdaki uygulamaya bakın)
- Bir yönetici başlatmak için yeni bir `ctrl.NewFromComponentConfig()` işlevi sağlamak
- `ComponentConfig` türlerini kolayca yazmak için gömülebilir bir `ControllerManagerConfiguration` türü sağlamak
- Müşteriler için geçişi kolaylaştırmak amacıyla bir `DefaultControllerConfig` sağlamak

### Hedef Dışı/Gelecek Çalışmalar

- `kubebuilder` uygulaması ve tasarımı başka bir PR'da
- Varsayılan `controller-runtime` uygulamasını değiştirmek
- `ComponentConfig` nesnesini dinamik olarak yeniden yüklemek
- `bayraklar` arayüzü ve geçersiz kılmalar sağlamak

## Teklif

`ctrl.Manager`, `ComponentConfig` benzeri nesnelerden yapılandırmaları yüklemeyi desteklemelidir.
Bu nesne için belirli yapılandırma parametreleri için getter'lar içeren bir arayüz oluşturulacaktır.

Mevcut `ctrl.NewManager`'ı kırmadan, `manager.go` yeni bir işlev, `NewFromComponentConfig()` açığa çıkarabilir. Bu işlev, getter'ları döngüye alarak iç `ctrl.Options{}`'ı doldurabilir ve bunu `New()`'a geçirebilir.

```golang
//pkg/manager/manager.go

// ManagerConfiguration, ControllerRuntime için ComponentConfig nesnesinin desteklemesi gerekenleri tanımlar
type ManagerConfiguration interface {
	GetSyncPeriod() *time.Duration

	GetLeaderElection() bool
	GetLeaderElectionNamespace() string
	GetLeaderElectionID() string

	GetLeaseDuration() *time.Duration
	GetRenewDeadline() *time.Duration
	GetRetryPeriod() *time.Duration

	GetNamespace() string
	GetMetricsBindAddress() string
	GetHealthProbeBindAddress() string

	GetReadinessEndpointName() string
	GetLivenessEndpointName() string

	GetPort() int
	GetHost() string

	GetCertDir() string
}

func NewFromComponentConfig(config *rest.Config, scheme *runtime.Scheme, filename string, managerconfig ManagerConfiguration) (Manager, error) {
	codecs := serializer.NewCodecFactory(scheme)
	 if err := decodeComponentConfigFileInto(codecs, filename, managerconfig); err != nil {

	}
	options := Options{}

	if scheme != nil {
		options.Scheme = scheme
	}

	// Getter'ları döngüye al
	if managerconfig.GetLeaderElection() {
		options.LeaderElection = managerconfig.GetLeaderElection()
	}
	// ...

	return New(config, options)
}
```

#### ComponentConfig Yükleme Sırası

![ComponentConfig Yükleme Sırası](/designs/images/component-config-load.png)

#### Gömülebilir ComponentConfig Türü

Controller yazarları için bunu kolaylaştırmak amacıyla `controller-runtime`, `config.ControllerConfiguration` türünden bir dizi tür açığa çıkarabilir. Bu türler, `k8s.io/apimachinery/pkg/apis/meta/v1`'in `TypeMeta` ve `ObjectMeta` için çalıştığı gibi gömülebilir. Bu türler `pkg/api/config/v1alpha1/types.go` içinde yer alabilir. Aşağıda `DefaultComponentConfig` türünün bir örnek uygulaması verilmiştir.

```golang
// pkg/api/config/v1alpha1/types.go
package v1alpha1

import (
	"time"

	configv1alpha1 "k8s.io/component-base/config/v1alpha1"
)

// ControllerManagerConfiguration, controller-runtime müşterileri için gömülü RuntimeConfiguration'ı tanımlar.
type ControllerManagerConfiguration struct {
	Namespace string `json:"namespace,omitempty"`

	SyncPeriod *time.Duration `json:"syncPeriod,omitempty"`

	LeaderElection configv1alpha1.LeaderElectionConfiguration `json:"leaderElection,omitempty"`

	MetricsBindAddress string `json:"metricsBindAddress,omitempty"`

	Health ControllerManagerConfigurationHealth `json:"health,omitempty"`

	Port *int   `json:"port,omitempty"`
	Host string `json:"host,omitempty"`

	CertDir string `json:"certDir,omitempty"`
}

// ControllerManagerConfigurationHealth, sağlık yapılandırmalarını tanımlar
type ControllerManagerConfigurationHealth struct {
	HealthProbeBindAddress string `json:"healthProbeBindAddress,omitempty"`

	ReadinessEndpointName string `json:"readinessEndpointName,omitempty"`
	LivenessEndpointName  string `json:"livenessEndpointName,omitempty"`
}
```

#### Varsayılan ComponentConfig Türü

`controller-runtime`'ın, her kontrolör veya uzantının kendi `ComponentConfig` türünü oluşturmasını gerektirmeden kullanılabilecek varsayılan bir `ComponentConfig` yapısına sahip olmasını sağlamak için `pkg/api/config/v1alpha1/types.go` içinde yer alabilecek bir `DefaultControllerConfiguration` türü oluşturabiliriz. Bu, kontrolör yazarlarının ek yapılandırmalarla kendi türlerini uygulamadan önce bu yapıyı kullanmalarına olanak tanır.

```golang
// pkg/api/config/v1alpha1/types.go
package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	configv1alpha1 "sigs.k8s.io/controller-runtime/pkg/apis/config/v1alpha1"
)

// DefaultControllerManagerConfiguration, DefaultControllerManagerConfigurations API'si için Şemayı tanımlar
type DefaultControllerManagerConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	Spec   configv1alpha1.ControllerManagerConfiguration   `json:"spec,omitempty"`
}
```

Bu, bir kontrolör yazarının json/yaml yapısını destekleyen herhangi bir yapılandırma ile bu yapıyı kullanmasına olanak tanır. Örneğin, bir kontrolör yazarı `Kind`'ını `FoobarControllerConfiguration` olarak tanımlayabilir ve aşağıdaki gibi tanımlayabilir.

```yaml
# config.yaml
apiVersion: config.somedomain.io/v1alpha1
kind: FoobarControllerManagerConfiguration
spec:
  port: 9443
  metricsBindAddress: ":8080"
  leaderElection:
	 leaderElect: false
```

Yukarıdaki yapılandırma ve `DefaultControllerManagerConfiguration` ile kontrolörü aşağıdaki şekilde başlatabiliriz.

```golang
mgr, err := ctrl.NewManagerFromComponentConfig(ctrl.GetConfigOrDie(), scheme, configname, &defaultv1alpha1.DefaultControllerManagerConfiguration{})
if err != nil {
	// ...
}
```

Yukarıdaki örnek, yapılandırmayı yüklemek için dosya adını ve belirli serileştiriciyi almak için `scheme`'i kullanır, örneğin `serializer.NewCodecFactory(scheme)`. Bu, yapılandırmanın `runtime.Object` türüne ayrıştırılmasına ve `ManagerConfiguration` arayüzü olarak `ctrl.NewManagerFromComponentConfig()`'a geçirilmesine olanak tanır.

#### ComponentConfig ile Bayrak Kullanımı

Bu tasarım hala başlangıç `ComponentConfig` türünü ayarlamayı ve `ctrl.NewFromComponentConfig()`'a bir işaretçi geçirmeyi gerektirdiğinden, kontrolörünüz herhangi bir bayrak arayüzünü kullanabilir. Örneğin [`flag`](https://golang.org/pkg/flag/), [`pflag`](https://pkg.go.dev/github.com/spf13/pflag), [`flagnum`](https://pkg.go.dev/github.com/luci/luci-go/common/flag/flagenum) ve `ComponentConfig` türünde değerler ayarlayabilir ve işaretçiyi `ctrl.NewFromComponentConfig()`'a geçirebilir, aşağıdaki örneğe bakın.

```golang
leaderElect := true

config := &defaultv1alpha1.DefaultControllerManagerConfiguration{
	Spec: configv1alpha1.ControllerManagerConfiguration{
		LeaderElection: configv1alpha1.LeaderElectionConfiguration{
			LeaderElect: &leaderElect,
		},
	},
}
mgr, err := ctrl.NewManagerFromComponentConfig(ctrl.GetConfigOrDie(), scheme, configname, config)
if err != nil {
	// ...
}
```

#### Kubebuilder İskelet Örneği

Ayrı bir tasarımda genişletilmiş olarak _(oluşturulduğunda bağlantı)_ bu, kontrolör yazarlarının `ManagerConfiguration` arayüzünü uygulayan bir tür oluşturmasına olanak tanır. Aşağıda bunun nasıl göründüğüne dair bir örnek verilmiştir:

```golang
package config

import (
  "time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	configv1alpha1 "sigs.k8s.io/controller-runtime/pkg/apis/config/v1alpha1"
)

type ControllerNameConfigurationSpec struct {
	configv1alpha1.ControllerManagerConfiguration `json:",inline"`
}

type ControllerNameConfiguration struct {
  metav1.TypeMeta

  Spec ControllerNameConfigurationSpec `json:"spec"`
}
```

Bu özel `ComponentConfig` türünü kullanmak, `ctrl.NewFromComponentConfig()`'u yeni yapı ile değiştirmeyi gerektirir.

## Kullanıcı Hikayeleri

### Varsayılan tür ile `controller-runtime` kullanan Controller Yazarı

- `ConfigMap`'i bağla
- Yapılandırma adı ve `DefaultControllerManagerConfiguration` türü ile `ctrl.Manager`'ı `NewFromComponentConfig` ile başlat
- Özel kontrolör oluştur

### Özel tür ile `controller-runtime` kullanan Controller Yazarı

- `ComponentConfig` türünü uygula
- `ControllerManagerConfiguration` türünü göm
- `ConfigMap`'i bağla
- Yapılandırma adı ve `ComponentConfig` türü ile `ctrl.Manager`'ı `NewFromComponentConfig` ile başlat
- Özel kontrolör oluştur

### `kubebuilder` ile Controller Yazarı (kubebuilder için tbd teklif)

- `--component-config-name=XYZConfiguration` kullanarak `kubebuilder` projesini başlat
- Özel kontrolör oluştur

### Yapılandırma değişiklikleri olmadan Controller Kullanıcısı

_Kontrolörün manifestleri sağladığı varsayılarak_

- Kontrolörü kümeye uygula
- Özel kaynakları dağıt

### Yapılandırma değişiklikleri ile Controller Kullanıcısı

- _Önceki örnekten değişiklikler olmadan devam ederek_
- Değişiklikler için yeni bir `ConfigMap` oluştur
- `controller-runtime` podunu yeni `ConfigMap`'i kullanacak şekilde değiştir
- Kontrolörü kümeye uygula
- Özel kaynakları dağıt

## Riskler ve Azaltmalar

- Bu, `controller-runtime` için çekirdek Yönetici başlatma işlemini değiştirmediğinden, oldukça düşük risklidir

## Alternatifler

* `NewFromComponentConfig()`, dosya adına dayalı olarak nesneyi diskten yükleyebilir ve `ComponentConfig` türünü doldurabilir.

## Uygulama Geçmişi

- [x] 02/19/2020: Bir sorun veya [topluluk toplantısında] öneri sunuldu
- [x] 02/24/2020: `controller-runtime`'a teklif sunuldu
- [x] 03/02/2020: Varsayılan `DefaultControllerManagerConfiguration` ile güncellendi
- [x] 03/04/2020: Gömülebilir `RuntimeConfig` ile güncellendi
- [x] 03/10/2020: Gömülebilir ad `ControllerManagerConfiguration` olarak güncellendi

<!-- Bağlantılar -->
[topluluk toplantısı]: https://docs.google.com/document/d/1Ih-2cgg1bUrLwLVTB9tADlPcVdgnuMNBGbUl4D-0TIk
