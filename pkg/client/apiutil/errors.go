/*
2023 Kubernetes Yazarları tarafından telif hakkı saklıdır.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN.
Lisans kapsamındaki izinler ve sınırlamalar hakkında daha fazla bilgi için Lisansı inceleyin.
*/

package apiutil

import (
	"fmt"
	"sort"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ErrResourceDiscoveryFailed, RESTMapper bazı GroupVersion'lar için desteklenen kaynakları keşfedemezse döndürülür.
// Karşılaşılan hataları sarar, "NotFound" hataları meta.NoResourceMatchError ile değiştirilir,
// meta.IsNoMatchError() kullanarak desteklenmeyen API'leri kontrol eden kodlarla geriye dönük uyumluluk için.
type ErrResourceDiscoveryFailed map[schema.GroupVersion]error

// Error, error arayüzünü uygular.
func (e *ErrResourceDiscoveryFailed) Error() string {
	altHatalar := []string{}
	for k, v := range *e {
		altHatalar = append(altHatalar, fmt.Sprintf("%s: %v", k, v))
	}
	sort.Strings(altHatalar)
	return fmt.Sprintf("sunucu API'lerinin tam listesini almak mümkün değil: %s", strings.Join(altHatalar, ", "))
}

// Unwrap, alt hataları döndürür.
func (e *ErrResourceDiscoveryFailed) Unwrap() []error {
	altHatalar := []error{}
	for gv, err := range *e {
		if apierrors.IsNotFound(err) {
			err = &meta.NoResourceMatchError{PartialResource: gv.WithResource("")}
		}
		altHatalar = append(altHatalar, err)
	}
	return altHatalar
}
