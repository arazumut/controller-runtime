/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamındaki izinler ve sınırlamalar hakkında daha fazla bilgi için
Lisans'a bakınız.
*/

package client_test

import (
	"bytes"
	"io"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/examples/crd/pkg"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Client Suite")
}

var (
	testenv   *envtest.Environment
	cfg       *rest.Config
	clientset *kubernetes.Clientset

	// Testlerin kontrolcü ve istemci log mesajlarını incelemesi için kullanılır.
	log bytes.Buffer
)

var _ = BeforeSuite(func() {
	// Logları ginkgo çıktısına yönlendirir ve testlerin logları incelemesine izin verir.
	mw := io.MultiWriter(&log, GinkgoWriter)

	// Log mesajlarının kaynağını ayırt etmemize yardımcı olmak için önekler kullanın.
	// controller-runtime logf kullanır
	logf.SetLogger(zap.New(zap.WriteTo(mw), zap.UseDevMode(true)).WithName("logf"))
	// client-go logları klog kullanır
	klog.SetLogger(zap.New(zap.WriteTo(mw), zap.UseDevMode(true)).WithName("klog"))

	testenv = &envtest.Environment{CRDDirectoryPaths: []string{"./testdata"}}

	var err error
	cfg, err = testenv.Start()
	Expect(err).NotTo(HaveOccurred())

	clientset, err = kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())

	Expect(pkg.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	Expect(testenv.Stop()).To(Succeed())
})
