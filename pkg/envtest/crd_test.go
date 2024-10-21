/*
2022 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği olmadıkça,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
herhangi bir garanti veya koşul olmaksızın.
Lisans kapsamındaki izinleri ve sınırlamaları yöneten özel dil için
Lisans'a bakınız.
*/

package envtest

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/sets"
)

var _ = Describe("Test", func() {
	Describe("readCRDFiles", func() {
		It("farklı dizinlerden gelen dosyaları karıştırmamalı", func() {
			opt := CRDInstallOptions{
				Paths: []string{
					"testdata/crds",
					"testdata/crdv1_original",
				},
			}
			err := readCRDFiles(&opt)
			Expect(err).NotTo(HaveOccurred())

			expectedCRDs := sets.NewString(
				"frigates.ship.example.com",
				"configs.foo.example.com",
				"drivers.crew.example.com",
			)

			foundCRDs := sets.NewString()
			for _, crd := range opt.CRDs {
				foundCRDs.Insert(crd.Name)
			}

			Expect(expectedCRDs).To(Equal(foundCRDs))
		})
	})
})
