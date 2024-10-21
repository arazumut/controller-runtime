/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
haklar ve sınırlamalar için Lisansa bakınız.
*/

package controllertest

import (
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ runtime.Object = &HataTipi{}

// HataTipi, runtime.Object'i uygular ancak herhangi bir şemada kayıtlı değildir ve bu nedenle testlerde hatalara neden olmalıdır.
type HataTipi struct{}

// GetObjectKind, runtime.Object'i uygular.
func (HataTipi) GetObjectKind() schema.ObjectKind { return nil }

// DeepCopyObject, runtime.Object'i uygular.
func (HataTipi) DeepCopyObject() runtime.Object { return nil }

var _ workqueue.TypedRateLimitingInterface[reconcile.Request] = &Kuyruk{}

// Kuyruk, test için oran sınırlaması olmayan bir kuyruk olarak bir Oran Sınırlama kuyruğunu uygular.
// Bu, bir Oran Sınırlama kuyruğu kullanan işlevlerin öğeleri kuyruğa senkronize olarak eklemesine yardımcı olur.
type Kuyruk = TipKuyruk[reconcile.Request]

// TipKuyruk, test için oran sınırlaması olmayan bir kuyruk olarak bir Oran Sınırlama kuyruğunu uygular.
// Bu, bir Oran Sınırlama kuyruğu kullanan işlevlerin öğeleri kuyruğa senkronize olarak eklemesine yardımcı olur.
type TipKuyruk[istek comparable] struct {
	workqueue.TypedInterface[istek]
	EklenenOranSınırlıKilit sync.Mutex
	EklenenOranSınırlı      []any
}

// AddAfter, RateLimitingInterface'i uygular.
func (q *TipKuyruk[istek]) AddAfter(item istek, duration time.Duration) {
	q.Add(item)
}

// AddRateLimited, RateLimitingInterface'i uygular.
func (q *TipKuyruk[istek]) AddRateLimited(item istek) {
	q.EklenenOranSınırlıKilit.Lock()
	q.EklenenOranSınırlı = append(q.EklenenOranSınırlı, item)
	q.EklenenOranSınırlıKilit.Unlock()
	q.Add(item)
}

// Forget, RateLimitingInterface'i uygular.
func (q *TipKuyruk[istek]) Forget(item istek) {}

// NumRequeues, RateLimitingInterface'i uygular.
func (q *TipKuyruk[istek]) NumRequeues(item istek) int {
	return 0
}
