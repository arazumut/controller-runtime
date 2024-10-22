/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisansa bakınız.
*/

package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// RequestLatency, kabul isteklerinin işlenme gecikmesinin histogramıdır.
	RequestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "controller_runtime_webhook_latency_seconds",
			Help: "Kabul isteklerinin işlenme gecikmesinin histogramı",
		},
		[]string{"webhook"},
	)

	// RequestTotal, toplam işlenen kabul isteklerinin sayacıdır.
	RequestTotal = func() *prometheus.CounterVec {
		return prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "controller_runtime_webhook_requests_total",
				Help: "HTTP durum koduna göre toplam kabul istek sayısı.",
			},
			[]string{"webhook", "code"},
		)
	}()

	// RequestInFlight, uçuş halindeki kabul isteklerinin göstergesidir.
	RequestInFlight = func() *prometheus.GaugeVec {
		return prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "controller_runtime_webhook_requests_in_flight",
				Help: "Şu anda hizmet verilen kabul isteklerinin sayısı.",
			},
			[]string{"webhook"},
		)
	}()
)

func init() {
	metrics.Registry.MustRegister(RequestLatency, RequestTotal, RequestInFlight)
}

// InstrumentedHook, verilen webhook üzerine bazı enstrümantasyon ekler.
func InstrumentedHook(path string, hookRaw http.Handler) http.Handler {
	lbl := prometheus.Labels{"webhook": path}

	lat := RequestLatency.MustCurryWith(lbl)
	cnt := RequestTotal.MustCurryWith(lbl)
	gge := RequestInFlight.With(lbl)

	// En olası HTTP durum kodlarını başlat.
	cnt.WithLabelValues("200")
	cnt.WithLabelValues("500")

	return promhttp.InstrumentHandlerDuration(
		lat,
		promhttp.InstrumentHandlerCounter(
			cnt,
			promhttp.InstrumentHandlerInFlight(gge, hookRaw),
		),
	)
}
