/*
2022 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisansa uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
yetkiler ve sınırlamalar için Lisansa bakınız.
*/

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// ReadCertificateTotal, toplam sertifika okuma sayısını tutan bir Prometheus sayaç metrikidir.
	ReadCertificateTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "certwatcher_read_certificate_total",
		Help: "Toplam sertifika okuma sayısı",
	})

	// ReadCertificateErrors, sertifika okuma hatalarının toplam sayısını tutan bir Prometheus sayaç metrikidir.
	ReadCertificateErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "certwatcher_read_certificate_errors_total",
		Help: "Toplam sertifika okuma hatası sayısı",
	})
)

func init() {
	metrics.Registry.MustRegister(
		ReadCertificateTotal,
		ReadCertificateErrors,
	)
}
