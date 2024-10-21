/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisansa bakınız.
*/

package recorder_test

import (
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/internal/recorder"
)

var _ = Describe("recorder.Provider", func() {
	makeBroadcaster := func() (record.EventBroadcaster, bool) { return record.NewBroadcaster(), true }
	Describe("NewProvider", func() {
		It("bir sağlayıcı örneği ve nil hata döndürmelidir.", func() {
			provider, err := recorder.NewProvider(cfg, httpClient, scheme.Scheme, logr.Discard(), makeBroadcaster)
			Expect(provider).NotTo(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})

		It("istemciyi başlatmada hata oluşursa bir hata döndürmelidir.", func() {
			// Yapılandırmayı geçersiz kıl
			cfg1 := *cfg
			cfg1.Host = "geçersiz host"
			_, err := recorder.NewProvider(&cfg1, httpClient, scheme.Scheme, logr.Discard(), makeBroadcaster)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("istemciyi başlatmada hata oluştu"))
		})
	})
	Describe("GetEventRecorder", func() {
		It("bir kaydedici örneği döndürmelidir.", func() {
			provider, err := recorder.NewProvider(cfg, httpClient, scheme.Scheme, logr.Discard(), makeBroadcaster)
			Expect(err).NotTo(HaveOccurred())

			recorder := provider.GetEventRecorderFor("test")
			Expect(recorder).NotTo(BeNil())
		})
	})
})
