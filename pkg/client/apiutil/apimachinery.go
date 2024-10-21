/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa gereği veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen özel haklar ve
sınırlamalar için Lisansa bakın.
*/

// Paket apiutil, RESTMapper'lar ve ham REST istemcileri oluşturma,
// ve bir nesnenin GVK'sını çıkarma gibi ham Kubernetes API makineleriyle
// çalışmak için yardımcı araçlar içerir.
package apiutil

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

var (
	protobufScheme     = runtime.NewScheme()
	protobufSchemeLock sync.RWMutex
)

func init() {
	// Şu anda yalnızca Protokol Tamponlarını uygulayan yerleşik kaynaklar için etkinleştirilmiştir.
	// Özel kaynaklar için, CRD'ler Protokol Tamponlarını destekleyemez ancak Birleştirilmiş API destekleyebilir.
	// Belgeye bakın: https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#advanced-features-and-flexibility
	if err := clientgoscheme.AddToScheme(protobufScheme); err != nil {
		panic(err)
	}
}

// AddToProtobufScheme, protobufScheme'e verilen SchemeBuilder'ı ekler,
// bu, protobuf'u destekleyen ek türler olmalıdır.
func AddToProtobufScheme(addToScheme func(*runtime.Scheme) error) error {
	protobufSchemeLock.Lock()
	defer protobufSchemeLock.Unlock()
	return addToScheme(protobufScheme)
}

// IsObjectNamespaced, nesnenin ad alanı kapsamlı olup olmadığını döndürür.
// Yapılandırılmamış nesneler için gvk nesnenin kendisinden bulunur.
func IsObjectNamespaced(obj runtime.Object, scheme *runtime.Scheme, restmapper meta.RESTMapper) (bool, error) {
	gvk, err := GVKForObject(obj, scheme)
	if err != nil {
		return false, err
	}

	return IsGVKNamespaced(gvk, restmapper)
}

// IsGVKNamespaced, sağlanan GVK'ya sahip nesnenin ad alanı kapsamlı olup olmadığını döndürür.
func IsGVKNamespaced(gvk schema.GroupVersionKind, restmapper meta.RESTMapper) (bool, error) {
	// Tam GVK kullanarak RESTMapping'i alın. Sürümü hariç tutarsak, Sürüm seti
	// mevcutsa önbelleğe alınmış Grup kullanılarak doldurulacaktır. Bu, çalışma zamanında kaydedilen CRD'lerin yeni Sürümlerini güncellerken hatalara yol açabilir.
	restmapping, err := restmapper.RESTMapping(schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}, gvk.Version)
	if err != nil {
		return false, fmt.Errorf("restmapping alınamadı: %w", err)
	}

	scope := restmapping.Scope.Name()
	if scope == "" {
		return false, errors.New("kapsam tanımlanamıyor, boş kapsam döndürüldü")
	}

	if scope != meta.RESTScopeNameRoot {
		return true, nil
	}
	return false, nil
}

// GVKForObject, verilen nesneyle ilişkili GroupVersionKind'ı bulur, eğer yalnızca tek bir GVK varsa.
func GVKForObject(obj runtime.Object, scheme *runtime.Scheme) (schema.GroupVersionKind, error) {
	// TODO: Bunu keyfi kapsayıcı türlere genelleştirmek istiyor muyuz?
	// Sanırım bunun için genelleştirilmiş bir form şeması veya bir şeyler gerekecek.
	// Ne yazık ki, varsayılan olarak çalışan güvenilir bir "GetGVK" arayüzü yok
	// doldurulmamış statik türler ve doldurulmuş "dinamik" türler
	// (yapılandırılmamış, kısmi, vb.)

	// KısmiObjectMetadata'yı kontrol edin, bu yapılandırılmamışa benzer, ancak ObjectKinds tarafından ele alınmaz
	_, isPartial := obj.(*metav1.PartialObjectMetadata)
	_, isPartialList := obj.(*metav1.PartialObjectMetadataList)
	if isPartial || isPartialList {
		// nesneyi tanımak için GVK'nın doldurulmuş olmasını gerektiriyoruz
		gvk := obj.GetObjectKind().GroupVersionKind()
		if len(gvk.Kind) == 0 {
			return schema.GroupVersionKind{}, runtime.NewMissingKindErr("yapılandırılmamış nesnenin türü yok")
		}
		if len(gvk.Version) == 0 {
			return schema.GroupVersionKind{}, runtime.NewMissingVersionErr("yapılandırılmamış nesnenin sürümü yok")
		}
		return gvk, nil
	}

	// Verilen şemayı kullanarak nesne için tüm GVK'ları alın.
	gvks, isUnversioned, err := scheme.ObjectKinds(obj)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	if isUnversioned {
		return schema.GroupVersionKind{}, fmt.Errorf("sürümden bağımsız tür için grup-sürüm-tür oluşturulamıyor %T", obj)
	}

	switch {
	case len(gvks) < 1:
		// Nesnenin GVK'sı yoksa, nesne şemaya kaydedilmemiş olabilir.
		// veya geçerli bir nesne değil.
		return schema.GroupVersionKind{}, fmt.Errorf("Go türü ile ilişkili GroupVersionKind yok %T, tür şemaya kaydedildi mi?", obj)
	case len(gvks) > 1:
		err := fmt.Errorf("Go türü ile ilişkili birden fazla GroupVersionKind var %T Şema içinde, bu, bir türün aynı anda birden fazla GVK için kaydedildiğinde olabilir", obj)

		// Nesne için birden fazla GVK bulduk.
		currentGVK := obj.GetObjectKind().GroupVersionKind()
		if !currentGVK.Empty() {
			// Temel nesnenin bir GVK'sı varsa, kullanmadan önce GVK listesindeki olup olmadığını kontrol edin.
			for _, gvk := range gvks {
				if gvk == currentGVK {
					return gvk, nil
				}
			}

			return schema.GroupVersionKind{}, fmt.Errorf(
				"%w: nesnenin sağlanan GroupVersionKind'ı %q Şema'nın listesinde bulunamadı; birini tahmin etmeyi reddediyor: %q", err, currentGVK, gvks)
		}

		// Bu yalnızca metav1.XYZ gibi şeyler için tetiklenmelidir --
		// normal sürümlü türler iyi olmalıdır.
		//
		// Daha fazla bilgi için https://github.com/kubernetes-sigs/controller-runtime/issues/362 adresine bakın.
		return schema.GroupVersionKind{}, fmt.Errorf(
			"%w: arayanlar, tür kayıtlarını yalnızca bir kez kaydedilecek şekilde düzeltebilir veya nesne için kullanılacak GroupVersionKind'ı belirtebilir; birini tahmin etmeyi reddediyor: %q", err, gvks)
	default:
		// Herhangi bir başka durumda, nesne için tek bir GVK bulduk.
		return gvks[0], nil
	}
}

// RESTClientForGVK, verilen GroupVersionKind ile ilişkili kaynağa erişebilen yeni bir rest.Interface oluşturur.
// REST istemcisi, ayarlanmışsa baseConfig'ten müzakere edilmiş serileştiriciyi kullanacak şekilde yapılandırılacaktır, aksi takdirde varsayılan bir serileştirici ayarlanacaktır.
func RESTClientForGVK(gvk schema.GroupVersionKind, isUnstructured bool, baseConfig *rest.Config, codecs serializer.CodecFactory, httpClient *http.Client) (rest.Interface, error) {
	if httpClient == nil {
		return nil, fmt.Errorf("httpClient boş olmamalıdır, bir istemci oluşturmak için rest.HTTPClientFor(c) kullanmayı düşünün")
	}
	return rest.RESTClientForConfigAndClient(createRestConfig(gvk, isUnstructured, baseConfig, codecs), httpClient)
}

// createRestConfig, temel yapılandırmayı kopyalar ve yeni bir dinlenme yapılandırması için gerekli alanları günceller.
func createRestConfig(gvk schema.GroupVersionKind, isUnstructured bool, baseConfig *rest.Config, codecs serializer.CodecFactory) *rest.Config {
	gv := gvk.GroupVersion()

	cfg := rest.CopyConfig(baseConfig)
	cfg.GroupVersion = &gv
	if gvk.Group == "" {
		cfg.APIPath = "/api"
	} else {
		cfg.APIPath = "/apis"
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = rest.DefaultKubernetesUserAgent()
	}
	// TODO: Uzun vadede, bunun gerçekten doğru olduğundan emin olmak için keşif veya başka bir şey kontrol etmek istiyoruz.
	if cfg.ContentType == "" && !isUnstructured {
		protobufSchemeLock.RLock()
		if protobufScheme.Recognizes(gvk) {
			cfg.ContentType = runtime.ContentTypeProtobuf
		}
		protobufSchemeLock.RUnlock()
	}

	if isUnstructured {
		// Nesne yapılandırılmamışsa, client-go dinamik serileştiriciyi kullanırız.
		cfg = dynamic.ConfigFor(cfg)
	} else {
		cfg.NegotiatedSerializer = serializerWithTargetZeroingDecode{NegotiatedSerializer: serializer.WithoutConversionCodecFactory{CodecFactory: codecs}}
	}

	return cfg
}

type serializerWithTargetZeroingDecode struct {
	runtime.NegotiatedSerializer
}

func (s serializerWithTargetZeroingDecode) DecoderToVersion(serializer runtime.Decoder, r runtime.GroupVersioner) runtime.Decoder {
	return targetZeroingDecoder{upstream: s.NegotiatedSerializer.DecoderToVersion(serializer, r)}
}

type targetZeroingDecoder struct {
	upstream runtime.Decoder
}

func (t targetZeroingDecoder) Decode(data []byte, defaults *schema.GroupVersionKind, into runtime.Object) (runtime.Object, *schema.GroupVersionKind, error) {
	zero(into)
	return t.upstream.Decode(data, defaults, into)
}

// zero, bir işaretçinin değerini sıfırlar.
func zero(x interface{}) {
	if x == nil {
		return
	}
	res := reflect.ValueOf(x).Elem()
	res.Set(reflect.Zero(res.Type()))
}
