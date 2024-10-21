/*
2020 Kubernetes Yazarları tarafından oluşturulmuştur.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VERİLMEKSİZİN, açık veya zımni olarak.
Lisans kapsamındaki izinler ve sınırlamalar hakkında daha fazla bilgi için
Lisans'a bakınız.
*/

package cluster

import (
	"context"
	"net/http"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"

	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	intrec "sigs.k8s.io/controller-runtime/pkg/internal/recorder"
)

// cluster yapısı, küme ile ilgili yapılandırma ve bileşenleri içerir.
type cluster struct {
	config           *rest.Config        // apiserver ile konuşmak için kullanılan rest.config. Gerekli.
	httpClient       *http.Client        // HTTP istemcisi
	scheme           *runtime.Scheme     // Şema
	cache            cache.Cache         // Önbellek
	client           client.Client       // İstemci
	apiReader        client.Reader       // API sunucusuna istek yapacak okuyucu, önbelleğe değil.
	fieldIndexes     client.FieldIndexer // Alan dizinleyici
	recorderProvider *intrec.Provider    // Olay kaydedici sağlayıcı
	mapper           meta.RESTMapper     // Kaynakları tür ve sürüme göre eşleyen haritalayıcı
	logger           logr.Logger         // Günlükleyici
}

// GetConfig, yapılandırmayı döndürür.
func (c *cluster) GetConfig() *rest.Config {
	return c.config
}

// GetHTTPClient, HTTP istemcisini döndürür.
func (c *cluster) GetHTTPClient() *http.Client {
	return c.httpClient
}

// GetClient, istemciyi döndürür.
func (c *cluster) GetClient() client.Client {
	return c.client
}

// GetScheme, şemayı döndürür.
func (c *cluster) GetScheme() *runtime.Scheme {
	return c.scheme
}

// GetFieldIndexer, alan dizinleyiciyi döndürür.
func (c *cluster) GetFieldIndexer() client.FieldIndexer {
	return c.fieldIndexes
}

// GetCache, önbelleği döndürür.
func (c *cluster) GetCache() cache.Cache {
	return c.cache
}

// GetEventRecorderFor, belirtilen ad için olay kaydediciyi döndürür.
func (c *cluster) GetEventRecorderFor(name string) record.EventRecorder {
	return c.recorderProvider.GetEventRecorderFor(name)
}

// GetRESTMapper, REST haritalayıcıyı döndürür.
func (c *cluster) GetRESTMapper() meta.RESTMapper {
	return c.mapper
}

// GetAPIReader, API okuyucusunu döndürür.
func (c *cluster) GetAPIReader() client.Reader {
	return c.apiReader
}

// GetLogger, günlükleyiciyi döndürür.
func (c *cluster) GetLogger() logr.Logger {
	return c.logger
}

// Start, küme bileşenlerini başlatır.
func (c *cluster) Start(ctx context.Context) error {
	defer c.recorderProvider.Stop(ctx)
	return c.cache.Start(ctx)
}
