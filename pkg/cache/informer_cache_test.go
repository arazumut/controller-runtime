/*
2021 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
yetkiler ve sınırlamalar için Lisansa bakınız.
*/

package cache_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var _ = Describe("informerCache", func() {
	It("Lider Seçimi gerektirmemeli", func() {
		cfg := &rest.Config{}

		httpClient, err := rest.HTTPClientFor(cfg)
		Expect(err).ToNot(HaveOccurred())
		mapper, err := apiutil.NewDynamicRESTMapper(cfg, httpClient)
		Expect(err).ToNot(HaveOccurred())

		c, err := cache.New(cfg, cache.Options{Mapper: mapper})
		Expect(err).ToNot(HaveOccurred())

		leaderElectionRunnable, ok := c.(manager.LeaderElectionRunnable)
		Expect(ok).To(BeTrue())
		Expect(leaderElectionRunnable.NeedLeaderElection()).To(BeFalse())
	})
})
