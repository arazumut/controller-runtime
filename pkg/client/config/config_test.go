/*
2019 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
haklar ve sınırlamalar için Lisansa bakın.
*/

package config

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type testDurumu struct {
	metin             string
	bağlam            string
	kubeconfigBayrağı string
	kubeconfigÇevre   []string
	istenenHost       string
}

var _ = Describe("Config", func() {

	var dizin string

	origRecommendedHomeFile := clientcmd.RecommendedHomeFile

	BeforeEach(func() {
		// test durumu için geçici dizin oluştur
		var err error
		dizin, err = os.MkdirTemp("", "cr-test")
		Expect(err).NotTo(HaveOccurred())

		// $HOME/.kube/config dosyasını geçersiz kıl
		clientcmd.RecommendedHomeFile = filepath.Join(dizin, ".kubeconfig")
	})

	AfterEach(func() {
		os.Unsetenv(clientcmd.RecommendedConfigPathEnvVar)
		kubeconfig = ""
		clientcmd.RecommendedHomeFile = origRecommendedHomeFile

		err := os.RemoveAll(dizin)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("GetConfigWithContext", func() {
		testleriTanımla := func(testDurumları []testDurumu) {
			for _, testDurumu := range testDurumları {
				tc := testDurumu
				It(tc.metin, func() {
					// global ve çevre yapılandırmalarını ayarla
					yapılandırmalarıAyarla(tc, dizin)

					// testi çalıştır
					cfg, err := GetConfigWithContext(tc.bağlam)
					Expect(err).NotTo(HaveOccurred())
					Expect(cfg.Host).To(Equal(tc.istenenHost))
				})
			}
		}

		Context("kubeconfig dosyaları mevcut değilken", func() {
			It("başarısız olmalı", func() {
				err := os.Unsetenv(clientcmd.RecommendedConfigPathEnvVar)
				Expect(err).NotTo(HaveOccurred())

				cfg, err := GetConfigWithContext("")
				Expect(cfg).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})

		Context("küme içindeyken", func() {
			kubeconfigDosyaları := map[string]string{
				"kubeconfig-multi-context": genKubeconfig("from-multi-env-1", "from-multi-env-2"),
				".kubeconfig":              genKubeconfig("from-home"),
			}
			BeforeEach(func() {
				err := dosyalarıOluştur(kubeconfigDosyaları, dizin)
				Expect(err).NotTo(HaveOccurred())

				// küme içi yapılandırma yükleyicisini geçersiz kıl
				loadInClusterConfig = func() (*rest.Config, error) {
					return &rest.Config{Host: "from-in-cluster"}, nil
				}
			})
			AfterEach(func() { loadInClusterConfig = rest.InClusterConfig })

			testDurumları := []testDurumu{
				{
					metin:           "çevre değişkenini küme içi yapılandırmanın üzerinde tercih etmeli",
					kubeconfigÇevre: []string{"kubeconfig-multi-context"},
					istenenHost:     "from-multi-env-1",
				},
				{
					metin:       "önerilen ev dosyasının üzerinde küme içi yapılandırmayı tercih etmeli",
					istenenHost: "from-in-cluster",
				},
			}
			testleriTanımla(testDurumları)
		})

		Context("küme dışındayken", func() {
			kubeconfigDosyaları := map[string]string{
				"kubeconfig-flag":          genKubeconfig("from-flag"),
				"kubeconfig-multi-context": genKubeconfig("from-multi-env-1", "from-multi-env-2"),
				"kubeconfig-env-1":         genKubeconfig("from-env-1"),
				"kubeconfig-env-2":         genKubeconfig("from-env-2"),
				".kubeconfig":              genKubeconfig("from-home"),
			}
			BeforeEach(func() {
				err := dosyalarıOluştur(kubeconfigDosyaları, dizin)
				Expect(err).NotTo(HaveOccurred())
			})
			testDurumları := []testDurumu{
				{
					metin:             "--kubeconfig bayrağını kullanmalı",
					kubeconfigBayrağı: "kubeconfig-flag",
					istenenHost:       "from-flag",
				},
				{
					metin:           "çevre değişkenini kullanmalı",
					kubeconfigÇevre: []string{"kubeconfig-multi-context"},
					istenenHost:     "from-multi-env-1",
				},
				{
					metin:       "önerilen ev dosyasını kullanmalı",
					istenenHost: "from-home",
				},
				{
					metin:             "bayrağı çevre değişkeninin üzerinde tercih etmeli",
					kubeconfigBayrağı: "kubeconfig-flag",
					kubeconfigÇevre:   []string{"kubeconfig-multi-context"},
					istenenHost:       "from-flag",
				},
				{
					metin:           "çevre değişkenini önerilen ev dosyasının üzerinde tercih etmeli",
					kubeconfigÇevre: []string{"kubeconfig-multi-context"},
					istenenHost:     "from-multi-env-1",
				},
				{
					metin:           "bağlamı geçersiz kılmaya izin vermeli",
					bağlam:          "from-multi-env-2",
					kubeconfigÇevre: []string{"kubeconfig-multi-context"},
					istenenHost:     "from-multi-env-2",
				},
				{
					metin:           "çok değerli bir çevre değişkenini desteklemeli",
					bağlam:          "from-env-2",
					kubeconfigÇevre: []string{"kubeconfig-env-1", "kubeconfig-env-2"},
					istenenHost:     "from-env-2",
				},
			}
			testleriTanımla(testDurumları)
		})
	})
})

func yapılandırmalarıAyarla(tc testDurumu, dizin string) {
	// kubeconfig bayrak değerini ayarla
	if len(tc.kubeconfigBayrağı) > 0 {
		kubeconfig = filepath.Join(dizin, tc.kubeconfigBayrağı)
	}

	// KUBECONFIG çevre değeri ayarla
	if len(tc.kubeconfigÇevre) > 0 {
		kubeconfigÇevreYolları := []string{}
		for _, k := range tc.kubeconfigÇevre {
			kubeconfigÇevreYolları = append(kubeconfigÇevreYolları, filepath.Join(dizin, k))
		}
		os.Setenv(clientcmd.RecommendedConfigPathEnvVar, strings.Join(kubeconfigÇevreYolları, ":"))
	}
}

func dosyalarıOluştur(dosyalar map[string]string, dizin string) error {
	for yol, veri := range dosyalar {
		if err := os.WriteFile(filepath.Join(dizin, yol), []byte(veri), 0644); err != nil { //nolint:gosec
			return err
		}
	}
	return nil
}

func genKubeconfig(bağlamlar ...string) string {
	var sb strings.Builder
	sb.WriteString(`---
apiVersion: v1
kind: Config
clusters:
`)
	for _, ctx := range bağlamlar {
		sb.WriteString(`- cluster:
	server: ` + ctx + `
  name: ` + ctx + `
`)
	}
	sb.WriteString("contexts:\n")
	for _, ctx := range bağlamlar {
		sb.WriteString(`- context:
	cluster: ` + ctx + `
	user: ` + ctx + `
  name: ` + ctx + `
`)
	}

	sb.WriteString("users:\n")
	for _, ctx := range bağlamlar {
		sb.WriteString(`- name: ` + ctx + `
`)
	}
	sb.WriteString("preferences: {}\n")
	if len(bağlamlar) > 0 {
		sb.WriteString("current-context: " + bağlamlar[0] + "\n")
	}

	return sb.String()
}
