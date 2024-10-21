/*
2020 Kubernetes Yazarları tarafından oluşturulmuştur.

Apache Lisansı, Sürüm 2.0 ("Lisans") kapsamında lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
yetkiler ve sınırlamalar için Lisansa bakınız.
*/

package controller

import (
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/util/sets"
)

var nameLock sync.Mutex
var usedNames sets.String

func checkName(name string) error {
	nameLock.Lock()
	defer nameLock.Unlock()
	if usedNames == nil {
		usedNames = sets.NewString()
	}

	if usedNames.Has(name) {
		return fmt.Errorf("isim %s olan kontrolcü zaten mevcut. Kontrolcü isimleri benzersiz olmalıdır, aksi takdirde aynı metriklere rapor veren birden fazla kontrolcü olabilir", name)
	}

	usedNames.Insert(name)

	return nil
}
