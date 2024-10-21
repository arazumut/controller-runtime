/*
2021 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisans'a uygun olarak kullanabilirsiniz.
Lisans'ın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamındaki izin ve sınırlamalar hakkında daha fazla bilgi için
Lisans'a bakınız.
*/

package certwatcher_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
	"sigs.k8s.io/controller-runtime/pkg/certwatcher/metrics"
)

var _ = Describe("CertWatcher", func() {
	var _ = Describe("certwatcher New", func() {
		It("sertifika/anahtar olmadan hata vermeli", func() {
			_, err := certwatcher.New("", "")
			Expect(err).To(HaveOccurred())
		})
	})

	var (
		certPath string
		keyPath  string
	)

	var _ = Describe("certwatcher Start", func() {
		var (
			ctx       context.Context
			ctxCancel context.CancelFunc
			watcher   *certwatcher.CertWatcher
		)

		BeforeEach(func() {
			ctx, ctxCancel = context.WithCancel(context.Background())

			err := writeCerts(certPath, keyPath, "127.0.0.1")
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() error {
				for _, file := range []string{certPath, keyPath} {
					_, err := os.ReadFile(file)
					if err != nil {
						return err
					}
					continue
				}

				return nil
			}).Should(Succeed())

			watcher, err = certwatcher.New(certPath, keyPath)
			Expect(err).ToNot(HaveOccurred())
		})

		startWatcher := func() (done <-chan struct{}) {
			doneCh := make(chan struct{})
			go func() {
				defer GinkgoRecover()
				defer close(doneCh)
				Expect(watcher.Start(ctx)).To(Succeed())
			}()
			// ilk sertifikayı okuyana kadar bekle
			Eventually(func() error {
				err := watcher.ReadCertificate()
				return err
			}).Should(Succeed())
			return doneCh
		}

		It("ilk sertifika/anahtarı okumalı", func() {
			doneCh := startWatcher()

			ctxCancel()
			Eventually(doneCh, "4s").Should(BeClosed())
		})

		It("sertifika değiştiğinde currentCert'i yeniden yüklemeli", func() {
			doneCh := startWatcher()
			called := atomic.Int64{}
			watcher.RegisterCallback(func(crt tls.Certificate) {
				called.Add(1)
				Expect(crt.Certificate).ToNot(BeEmpty())
			})

			firstcert, _ := watcher.GetCertificate(nil)

			err := writeCerts(certPath, keyPath, "192.168.0.1")
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				secondcert, _ := watcher.GetCertificate(nil)
				first := firstcert.PrivateKey.(*rsa.PrivateKey)
				return first.Equal(secondcert.PrivateKey)
			}).ShouldNot(BeTrue())

			ctxCancel()
			Eventually(doneCh, "4s").Should(BeClosed())
			Expect(called.Load()).To(BeNumerically(">=", 1))
		})

		It("yeniden adlandırıldığında currentCert'i yeniden yüklemeli", func() {
			doneCh := startWatcher()
			called := atomic.Int64{}
			watcher.RegisterCallback(func(crt tls.Certificate) {
				called.Add(1)
				Expect(crt.Certificate).ToNot(BeEmpty())
			})

			firstcert, _ := watcher.GetCertificate(nil)

			err := writeCerts(certPath+".new", keyPath+".new", "192.168.0.2")
			Expect(err).ToNot(HaveOccurred())

			Expect(os.Link(certPath, certPath+".old")).To(Succeed())
			Expect(os.Rename(certPath+".new", certPath)).To(Succeed())

			Expect(os.Link(keyPath, keyPath+".old")).To(Succeed())
			Expect(os.Rename(keyPath+".new", keyPath)).To(Succeed())

			Eventually(func() bool {
				secondcert, _ := watcher.GetCertificate(nil)
				first := firstcert.PrivateKey.(*rsa.PrivateKey)
				return first.Equal(secondcert.PrivateKey)
			}).ShouldNot(BeTrue())

			ctxCancel()
			Eventually(doneCh, "4s").Should(BeClosed())
			Expect(called.Load()).To(BeNumerically(">=", 1))
		})

		Context("prometheus metrik read_certificate_total", func() {
			var readCertificateTotalBefore float64
			var readCertificateErrorsBefore float64

			BeforeEach(func() {
				readCertificateTotalBefore = testutil.ToFloat64(metrics.ReadCertificateTotal)
				readCertificateErrorsBefore = testutil.ToFloat64(metrics.ReadCertificateErrors)
			})

			It("başarılı sertifika okuma durumunda güncellenmeli", func() {
				doneCh := startWatcher()

				Eventually(func() error {
					readCertificateTotalAfter := testutil.ToFloat64(metrics.ReadCertificateTotal)
					if readCertificateTotalAfter != readCertificateTotalBefore+1.0 {
						return fmt.Errorf("metrik okunan sertifika toplamı beklenen: %v ve alınan: %v", readCertificateTotalBefore+1.0, readCertificateTotalAfter)
					}
					return nil
				}, "4s").Should(Succeed())

				ctxCancel()
				Eventually(doneCh, "4s").Should(BeClosed())
			})

			It("sertifika okuma hatalarında güncellenmeli", func() {
				doneCh := startWatcher()

				Eventually(func() error {
					readCertificateTotalAfter := testutil.ToFloat64(metrics.ReadCertificateTotal)
					if readCertificateTotalAfter != readCertificateTotalBefore+1.0 {
						return fmt.Errorf("metrik okunan sertifika toplamı beklenen: %v ve alınan: %v", readCertificateTotalBefore+1.0, readCertificateTotalAfter)
					}
					readCertificateTotalBefore = readCertificateTotalAfter
					return nil
				}, "4s").Should(Succeed())

				Expect(os.Remove(keyPath)).To(Succeed())

				// Not: Burada iki hata kontrol ediyoruz, çünkü os.Remove iki fsnotify olayı oluşturur: Chmod + Remove
				Eventually(func() error {
					readCertificateTotalAfter := testutil.ToFloat64(metrics.ReadCertificateTotal)
					if readCertificateTotalAfter != readCertificateTotalBefore+2.0 {
						return fmt.Errorf("metrik okunan sertifika toplamı beklenen: %v ve alınan: %v", readCertificateTotalBefore+2.0, readCertificateTotalAfter)
					}
					return nil
				}, "4s").Should(Succeed())
				Eventually(func() error {
					readCertificateErrorsAfter := testutil.ToFloat64(metrics.ReadCertificateErrors)
					if readCertificateErrorsAfter != readCertificateErrorsBefore+2.0 {
						return fmt.Errorf("metrik okunan sertifika hataları beklenen: %v ve alınan: %v", readCertificateErrorsBefore+2.0, readCertificateErrorsAfter)
					}
					return nil
				}, "4s").Should(Succeed())

				ctxCancel()
				Eventually(doneCh, "4s").Should(BeClosed())
			})
		})
	})
})

func writeCerts(certPath, keyPath, ip string) error {
	var priv interface{}
	var err error
	priv, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	keyUsage := x509.KeyUsageDigitalSignature
	if _, isRSA := priv.(*rsa.PrivateKey); isRSA {
		keyUsage |= x509.KeyUsageKeyEncipherment
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(1 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Kubernetes"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	template.IPAddresses = append(template.IPAddresses, net.ParseIP(ip))

	privkey := priv.(*rsa.PrivateKey)

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privkey.PublicKey, priv)
	if err != nil {
		return err
	}

	certOut, err := os.Create(certPath)
	if err != nil {
		return err
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return err
	}
	if err := certOut.Close(); err != nil {
		return err
	}

	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return err
	}
	return keyOut.Close()
}
