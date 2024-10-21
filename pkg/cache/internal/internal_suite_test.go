/*
2022 Kubernetes Yazarları tarafından oluşturulmuştur.

Apache Lisansı, Sürüm 2.0 ("Lisans") kapsamında lisanslanmıştır;
bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği olmadıkça,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen özel dildeki sınırlamalar ve
izinler için Lisansa bakınız.
*/

package internal

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestSource fonksiyonu, testlerin çalıştırılmasını sağlar.
func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cache Internal Suite")
}
