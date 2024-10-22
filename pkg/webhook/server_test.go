/*
2019 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN.
Lisans kapsamında izin verilen belirli dil kapsamındaki
haklar ve sınırlamalar için Lisansa bakın.
*/

package webhook_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"path"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var _ = Describe("Webhook Sunucusu", func() {
	var (
		ctx          context.Context
		ctxCancel    context.CancelFunc
		testHostPort string
		client       *http.Client
		server       webhook.Server
		servingOpts  envtest.WebhookInstallOptions
	)

	BeforeEach(func() {
		ctx, ctxCancel = context.WithCancel(context.Background())
		// Bireysel testlerde farklı şekilde kapatılır

		servingOpts = envtest.WebhookInstallOptions{}
		Expect(servingOpts.PrepWithoutInstalling()).To(Succeed())

		testHostPort = net.JoinHostPort(servingOpts.LocalServingHost, fmt.Sprintf("%d", servingOpts.LocalServingPort))

		// x509 sertifika havuzunu vb. kendimiz kurmamıza gerek kalmadan atla
		clientTransport, err := rest.TransportFor(&rest.Config{
			TLSClientConfig: rest.TLSClientConfig{CAData: servingOpts.LocalServingCAData},
		})
		Expect(err).NotTo(HaveOccurred())
		client = &http.Client{
			Transport: clientTransport,
		}

		server = webhook.NewServer(webhook.Options{
			Host:    servingOpts.LocalServingHost,
			Port:    servingOpts.LocalServingPort,
			CertDir: servingOpts.LocalServingCertDir,
		})
	})
	AfterEach(func() {
		Expect(servingOpts.Cleanup()).To(Succeed())
	})

	genericStartServer := func(f func(ctx context.Context)) (done <-chan struct{}) {
		doneCh := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			defer close(doneCh)
			f(ctx)
		}()
		// Sunucuyu başlatmak için testin başlamasını bekleyin
		Eventually(func() error {
			_, err := client.Get(fmt.Sprintf("https://%s/unservedpath", testHostPort))
			return err
		}).Should(Succeed())

		return doneCh
	}

	startServer := func() (done <-chan struct{}) {
		return genericStartServer(func(ctx context.Context) {
			Expect(server.Start(ctx)).To(Succeed())
		})
	}

	// TODO: httptest.Server ile tüm sunucu kurulumunu test etmek için iyi bir yol bulun

	Context("sunucu çalışırken", func() {
		PIt("istemci CA adını doğrulamalıdır", func() {

		})
		PIt("HTTP/2'yi desteklemelidir", func() {

		})

		// TODO: port varsayılanını vb. test etmek için iyi bir yol bulun
	})

	It("aynı yol tekrar kaydedilirse paniklemelidir", func() {
		server.Register("/somepath", &testHandler{})
		doneCh := startServer()

		Expect(func() { server.Register("/somepath", &testHandler{}) }).To(Panic())

		ctxCancel()
		Eventually(doneCh, "4s").Should(BeClosed())
	})

	Context("sunucu başlamadan önce yeni webhooks kaydedildiğinde", func() {
		It("istenen yolda bir webhook sunmalıdır", func() {
			server.Register("/somepath", &testHandler{})

			Expect(server.StartedChecker()(nil)).ToNot(Succeed())

			doneCh := startServer()

			Eventually(func() ([]byte, error) {
				resp, err := client.Get(fmt.Sprintf("https://%s/somepath", testHostPort))
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()
				return io.ReadAll(resp.Body)
			}).Should(Equal([]byte("gadzooks!")))

			Expect(server.StartedChecker()(nil)).To(Succeed())

			ctxCancel()
			Eventually(doneCh, "4s").Should(BeClosed())
		})
	})

	Context("sunucu başladıktan sonra webhooks kaydedildiğinde", func() {
		var (
			doneCh <-chan struct{}
		)
		BeforeEach(func() {
			doneCh = startServer()
		})
		AfterEach(func() {
			// Temizliğin gerçekleşmesini bekleyin
			ctxCancel()
			Eventually(doneCh, "4s").Should(BeClosed())
		})

		It("istenen yolda bir webhook sunmalıdır", func() {
			server.Register("/somepath", &testHandler{})
			resp, err := client.Get(fmt.Sprintf("https://%s/somepath", testHostPort))
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(io.ReadAll(resp.Body)).To(Equal([]byte("gadzooks!")))
		})
	})

	It("geçilen TLS yapılandırmalarına saygı göstermelidir", func() {
		var finalCfg *tls.Config
		tlsCfgFunc := func(cfg *tls.Config) {
			cfg.CipherSuites = []uint16{
				tls.TLS_AES_128_GCM_SHA256,
				tls.TLS_AES_256_GCM_SHA384,
			}
			cfg.MinVersion = tls.VersionTLS12
			// Yapılan değişikliklerden sonra cfg'yi test etmek için kaydedin
			finalCfg = cfg
		}
		server = webhook.NewServer(webhook.Options{
			Host:    servingOpts.LocalServingHost,
			Port:    servingOpts.LocalServingPort,
			CertDir: servingOpts.LocalServingCertDir,
			TLSOpts: []func(*tls.Config){
				tlsCfgFunc,
			},
		})
		server.Register("/somepath", &testHandler{})
		doneCh := genericStartServer(func(ctx context.Context) {
			Expect(server.Start(ctx)).To(Succeed())
		})

		Eventually(func() ([]byte, error) {
			resp, err := client.Get(fmt.Sprintf("https://%s/somepath", testHostPort))
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			return io.ReadAll(resp.Body)
		}).Should(Equal([]byte("gadzooks!")))
		Expect(finalCfg.MinVersion).To(Equal(uint16(tls.VersionTLS12)))
		Expect(finalCfg.CipherSuites).To(ContainElements(
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
		))

		ctxCancel()
		Eventually(doneCh, "4s").Should(BeClosed())
	})

	It("TLSOpts üzerinden GetCertificate'i tercih etmelidir", func() {
		var finalCfg *tls.Config
		finalCert, err := tls.LoadX509KeyPair(
			path.Join(servingOpts.LocalServingCertDir, "tls.crt"),
			path.Join(servingOpts.LocalServingCertDir, "tls.key"),
		)
		Expect(err).NotTo(HaveOccurred())
		finalGetCertificate := func(_ *tls.ClientHelloInfo) (*tls.Certificate, error) { //nolint:unparam
			return &finalCert, nil
		}
		server = &webhook.DefaultServer{Options: webhook.Options{
			Host:    servingOpts.LocalServingHost,
			Port:    servingOpts.LocalServingPort,
			CertDir: servingOpts.LocalServingCertDir,

			TLSOpts: []func(*tls.Config){
				func(cfg *tls.Config) {
					cfg.GetCertificate = finalGetCertificate
					cfg.MinVersion = tls.VersionTLS12
					// Yapılan değişikliklerden sonra cfg'yi test etmek için kaydedin
					finalCfg = cfg
				},
			},
		}}
		server.Register("/somepath", &testHandler{})
		doneCh := genericStartServer(func(ctx context.Context) {
			Expect(server.Start(ctx)).To(Succeed())
		})

		Eventually(func() ([]byte, error) {
			resp, err := client.Get(fmt.Sprintf("https://%s/somepath", testHostPort))
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			return io.ReadAll(resp.Body)
		}).Should(Equal([]byte("gadzooks!")))
		Expect(finalCfg.MinVersion).To(Equal(uint16(tls.VersionTLS12)))
		// Fonksiyonları doğrudan karşılaştıramayız, ancak işaretçilerini karşılaştırabiliriz
		if reflect.ValueOf(finalCfg.GetCertificate).Pointer() != reflect.ValueOf(finalGetCertificate).Pointer() {
			Fail("GetCertificate düzgün ayarlanmadı veya üzerine yazıldı")
		}
		cert, err := finalCfg.GetCertificate(nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(cert).To(BeEquivalentTo(&finalCert))

		ctxCancel()
		Eventually(doneCh, "4s").Should(BeClosed())
	})
})

type testHandler struct {
}

func (t *testHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if _, err := resp.Write([]byte("gadzooks!")); err != nil {
		panic("HTTP yanıtı yazılamadı!")
	}
}
