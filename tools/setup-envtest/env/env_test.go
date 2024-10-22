/*
2021 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği olmadıkça,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
yetkiler ve sınırlamalar için Lisansa bakınız.
*/

package env_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	. "sigs.k8s.io/controller-runtime/tools/setup-envtest/env"
	"sigs.k8s.io/controller-runtime/tools/setup-envtest/store"
	"sigs.k8s.io/controller-runtime/tools/setup-envtest/versions"
)

var _ = Describe("Env", func() {
	// Çoğu test e2e olarak workflows testi ile yapılır,
	// ancak burada test edilmesi daha kolay olan birkaç şey var.
	// Belki bazı testleri buraya taşımalıyız.
	var (
		env       *Env
		outBuffer *bytes.Buffer
	)
	BeforeEach(func() {
		outBuffer = new(bytes.Buffer)
		env = &Env{
			Out: outBuffer,
			Log: testLog,

			Store: &store.Store{
				// Boşluklar ve tırnak işaretlerini test etmek için
				Root: afero.NewBasePathFs(afero.NewMemMapFs(), "/kb's test store"),
			},

			// Bunlar kullanılmamalı, ama yine de
			NoDownload: true,
			FS:         afero.Afero{Fs: afero.NewMemMapFs()},
		}

		env.Version.MakeConcrete(versions.Concrete{
			Major: 1, Minor: 21, Patch: 3,
		})
		env.Platform.Platform = versions.Platform{
			OS: "linux", Arch: "amd64",
		}
	})

	Describe("yazdırma", func() {
		It("manuel bir yol varsa onu kullanmalı", func() {
			By("manuel bir yol kullanarak")
			Expect(env.PathMatches("/otherstore/1.21.4-linux-amd64")).To(BeTrue())

			By("bu yolun düzgün yazdırıldığını kontrol ederek")
			env.PrintInfo(PrintPath)
			Expect(outBuffer.String()).To(Equal("/otherstore/1.21.4-linux-amd64"))
		})

		Context("insan tarafından okunabilir bilgi olarak", func() {
			BeforeEach(func() {
				env.PrintInfo(PrintOverview)
			})

			It("sürümü içermeli", func() {
				Expect(outBuffer.String()).To(ContainSubstring("/kb's test store/k8s/1.21.3-linux-amd64"))
			})
			It("yolu içermeli", func() {
				Expect(outBuffer.String()).To(ContainSubstring("1.21.3"))
			})
			It("platformu içermeli", func() {
				Expect(outBuffer.String()).To(ContainSubstring("linux/amd64"))
			})

		})
		Context("sadece bir yol olarak", func() {
			It("sadece yolu yazdırmalı", func() {
				env.PrintInfo(PrintPath)
				Expect(outBuffer.String()).To(Equal(`/kb's test store/k8s/1.21.3-linux-amd64`))
			})
		})

		Context("çevre değişkenleri olarak", func() {
			BeforeEach(func() {
				env.PrintInfo(PrintEnv)
			})
			It("KUBEBUILDER_ASSETS'i ayarlamalı", func() {
				Expect(outBuffer.String()).To(HavePrefix("export KUBEBUILDER_ASSETS="))
			})
			It("dönüş yolunu tırnak içine almalı, boşluklar ve benzeri şeylerle başa çıkmak için tırnak işaretlerini kaçırmalı", func() {
				Expect(outBuffer.String()).To(HaveSuffix(`='/kb'"'"'s test store/k8s/1.21.3-linux-amd64'` + "\n"))
			})
		})
	})
})
