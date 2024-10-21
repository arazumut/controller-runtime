/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") kapsamında lisanslanmıştır;
bu dosyayı ancak Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa kapsamında gerekli olmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisans'a bakın.
*/

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// ReconcileTotal, her kontrolör için toplam uzlaştırma sayısını tutan bir prometheus sayaç metrikidir.
	// İki etiketi vardır: controller etiketi kontrolör adını ve result etiketi uzlaştırma sonucunu ifade eder.
	// Örneğin: başarı, hata, yeniden sıraya alma, yeniden sıraya alma sonrası.
	ReconcileTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "controller_runtime_reconcile_total",
		Help: "Her kontrolör için toplam uzlaştırma sayısı",
	}, []string{"controller", "result"})

	// ReconcileErrors, Uzlaştırıcıdan gelen toplam hata sayısını tutan bir prometheus sayaç metrikidir.
	ReconcileErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "controller_runtime_reconcile_errors_total",
		Help: "Her kontrolör için toplam uzlaştırma hatası sayısı",
	}, []string{"controller"})

	// TerminalReconcileErrors, Uzlaştırıcıdan gelen toplam terminal hata sayısını tutan bir prometheus sayaç metrikidir.
	TerminalReconcileErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "controller_runtime_terminal_reconcile_errors_total",
		Help: "Her kontrolör için toplam terminal uzlaştırma hatası sayısı",
	}, []string{"controller"})

	// ReconcilePanics, Uzlaştırıcıdan gelen toplam panik sayısını tutan bir prometheus sayaç metrikidir.
	ReconcilePanics = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "controller_runtime_reconcile_panics_total",
		Help: "Her kontrolör için toplam uzlaştırma panik sayısı",
	}, []string{"controller"})

	// ReconcileTime, uzlaştırmaların süresini takip eden bir prometheus metrikidir.
	ReconcileTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "controller_runtime_reconcile_time_seconds",
		Help: "Her kontrolör için uzlaştırma başına süre",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.15, 0.2, 0.25, 0.3, 0.35, 0.4, 0.45, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0,
			1.25, 1.5, 1.75, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5, 5, 6, 7, 8, 9, 10, 15, 20, 25, 30, 40, 50, 60},
	}, []string{"controller"})

	// WorkerCount, her kontrolör için eşzamanlı uzlaştırma sayısını tutan bir prometheus metrikidir.
	WorkerCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "controller_runtime_max_concurrent_reconciles",
		Help: "Her kontrolör için maksimum eşzamanlı uzlaştırma sayısı",
	}, []string{"controller"})

	// ActiveWorkers, her kontrolör için aktif işçi sayısını tutan bir prometheus metrikidir.
	ActiveWorkers = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "controller_runtime_active_workers",
		Help: "Her kontrolör için şu anda kullanılan işçi sayısı",
	}, []string{"controller"})
)

func init() {
	metrics.Registry.MustRegister(
		ReconcileTotal,
		ReconcileErrors,
		TerminalReconcileErrors,
		ReconcilePanics,
		ReconcileTime,
		WorkerCount,
		ActiveWorkers,
		// CPU, Bellek, dosya tanımlayıcı kullanımı gibi işlem metriklerini açığa çıkar.
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		// GC istatistikleri, bellek istatistikleri gibi Go çalışma zamanı metriklerini açığa çıkar.
		collectors.NewGoCollector(),
	)
}
