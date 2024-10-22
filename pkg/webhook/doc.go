/*
2018 Kubernetes Yazarları tarafından oluşturulmuştur.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN.
Lisans kapsamındaki izinler ve sınırlamalar hakkında daha fazla bilgi için Lisansı inceleyin.
*/

/*
Webhook paketi, bir webhook sunucusu oluşturmak ve başlatmak için yöntemler sağlar.

Şu anda yalnızca kabul webhooks'larını desteklemektedir. Yakın gelecekte CRD dönüşüm webhooks'larını destekleyecektir.
*/
package webhook

import (
	logf "sigs.k8s.io/controller-runtime/pkg/internal/log"
)

var log = logf.RuntimeLog.WithName("webhook")
