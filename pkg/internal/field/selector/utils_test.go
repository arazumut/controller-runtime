/*
2022 Kubernetes Yazarları tarafından oluşturulmuştur.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
bu yazılım Lisans kapsamında "OLDUĞU GİBİ" dağıtılmakta olup,
herhangi bir garanti veya koşul içermez.
Lisans kapsamındaki izinler ve sınırlamalar hakkında daha fazla bilgi için
Lisans'a bakınız.
*/

package selector_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/fields"

	. "sigs.k8s.io/controller-runtime/pkg/internal/field/selector"
)

var _ = Describe("RequiresExactMatch fonksiyonu", func() {

	It("Seçici her şeyi eşleştirdiğinde false döner", func() {
		requiresExactMatch := RequiresExactMatch(fields.Everything())
		Expect(requiresExactMatch).To(BeFalse())
	})

	It("Seçici hiçbir şeyi eşleştirmediğinde false döner", func() {
		requiresExactMatch := RequiresExactMatch(fields.Nothing())
		Expect(requiresExactMatch).To(BeFalse())
	})

	It("Seçici key!=val formunda olduğunda false döner", func() {
		requiresExactMatch := RequiresExactMatch(fields.ParseSelectorOrDie("key!=val"))
		Expect(requiresExactMatch).To(BeFalse())
	})

	It("Seçici key1==val1,key2==val2 formunda olduğunda true döner", func() {
		requiresExactMatch := RequiresExactMatch(fields.ParseSelectorOrDie("key1==val1,key2==val2"))
		Expect(requiresExactMatch).To(BeTrue())
	})

	It("Seçici key==val formunda olduğunda true döner", func() {
		requiresExactMatch := RequiresExactMatch(fields.ParseSelectorOrDie("key==val"))
		Expect(requiresExactMatch).To(BeTrue())
	})

	It("Seçici key=val formunda olduğunda true döner", func() {
		requiresExactMatch := RequiresExactMatch(fields.ParseSelectorOrDie("key=val"))
		Expect(requiresExactMatch).To(BeTrue())
	})
})
