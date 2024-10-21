/*
2014 Kubernetes Yazarları tarafından oluşturulmuştur.

Apache Lisansı, Sürüm 2.0 ("Lisans") altında lisanslanmıştır;
bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
yetkiler ve sınırlamalar için Lisansa bakınız.
*/

// healthz paketi, liveness ve readiness uç noktalarını desteklemek için yardımcı işlevler içerir.
// (sırasıyla healthz ve readyz olarak anılır).
//
// Bu paket, apiserver'ın healthz paketinden ( https://github.com/kubernetes/apiserver/tree/master/pkg/server/healthz )
// büyük ölçüde esinlenmiştir, ancak controller-runtime'ın tarzına uygun hale getirmek için bazı değişiklikler yapılmıştır.
//
// Ana giriş noktası Handler'dır -- bu, hem birleştirilmiş sağlık durumu hem de bireysel sağlık kontrol uç noktalarını sunar.
package healthz

import (
	logf "sigs.k8s.io/controller-runtime/pkg/internal/log"
)

var log = logf.RuntimeLog.WithName("healthz")
