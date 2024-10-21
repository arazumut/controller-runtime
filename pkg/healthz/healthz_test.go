/*
2014 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği olmadıkça,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMADAN, açık veya zımni.
Lisans kapsamındaki izinler ve sınırlamalar hakkında daha fazla bilgi için
Lisans'a bakınız.
*/

package healthz_test

import (
	"errors"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

const (
	contentType = "text/plain; charset=utf-8"
)

func requestTo(handler http.Handler, dest string) *httptest.ResponseRecorder {
	req, err := http.NewRequest("GET", dest, nil)
	Expect(err).NotTo(HaveOccurred())
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)

	return resp
}

var _ = Describe("Healthz Handler", func() {
	Describe("toplu endpoint", func() {
		It("tüm kontroller başarılı olursa sağlıklı dönmeli", func() {
			handler := &healthz.Handler{Checks: map[string]healthz.Checker{
				"ok1": healthz.Ping,
				"ok2": healthz.Ping,
			}}

			resp := requestTo(handler, "/")
			Expect(resp.Code).To(Equal(http.StatusOK))
		})

		It("en az bir kontrol başarısız olursa sağlıksız dönmeli", func() {
			handler := &healthz.Handler{Checks: map[string]healthz.Checker{
				"ok1": healthz.Ping,
				"bad1": func(req *http.Request) error {
					return errors.New("blech")
				},
			}}

			resp := requestTo(handler, "/")
			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
		})

		It("sağlık belirlerken hariç tutulan kontrolleri göz ardı etmeli", func() {
			handler := &healthz.Handler{Checks: map[string]healthz.Checker{
				"ok1": healthz.Ping,
				"bad1": func(req *http.Request) error {
					return errors.New("blech")
				},
			}}

			resp := requestTo(handler, "/?exclude=bad1")
			Expect(resp.Code).To(Equal(http.StatusOK))
		})

		It("var olmayan bir kontrolü hariç tutması istendiğinde sorun olmamalı", func() {
			handler := &healthz.Handler{Checks: map[string]healthz.Checker{
				"ok1": healthz.Ping,
				"ok2": healthz.Ping,
			}}

			resp := requestTo(handler, "/?exclude=nonexistant")
			Expect(resp.Code).To(Equal(http.StatusOK))
		})

		Context("?verbose=true ile ayrıntılı çıktı istendiğinde", func() {
			It("sağlıklı durumlar için ayrıntılı çıktı vermeli", func() {
				handler := &healthz.Handler{Checks: map[string]healthz.Checker{
					"ok1": healthz.Ping,
					"ok2": healthz.Ping,
				}}

				resp := requestTo(handler, "/?verbose=true")
				Expect(resp.Code).To(Equal(http.StatusOK))
				Expect(resp.Header().Get("Content-Type")).To(Equal(contentType))
				Expect(resp.Body.String()).To(Equal("[+]ok1 ok\n[+]ok2 ok\nhealthz kontrolü geçti\n"))
			})

			It("başarısızlıklar için ayrıntılı çıktı vermeli", func() {
				handler := &healthz.Handler{Checks: map[string]healthz.Checker{
					"ok1": healthz.Ping,
					"bad1": func(req *http.Request) error {
						return errors.New("blech")
					},
				}}

				resp := requestTo(handler, "/?verbose=true")
				Expect(resp.Code).To(Equal(http.StatusInternalServerError))
				Expect(resp.Header().Get("Content-Type")).To(Equal(contentType))
				Expect(resp.Body.String()).To(Equal("[-]bad1 başarısız: sebep gizlendi\n[+]ok1 ok\nhealthz kontrolü başarısız\n"))
			})
		})

		It("sağlıklı olduğunda ve ayrıntılı olarak belirtilmediğinde ayrıntısız çıktı vermeli", func() {
			handler := &healthz.Handler{Checks: map[string]healthz.Checker{
				"ok1": healthz.Ping,
				"ok2": healthz.Ping,
			}}

			resp := requestTo(handler, "/")
			Expect(resp.Header().Get("Content-Type")).To(Equal(contentType))
			Expect(resp.Body.String()).To(Equal("ok"))
		})

		It("bir kontrol başarısız olursa her zaman ayrıntılı olmalı", func() {
			handler := &healthz.Handler{Checks: map[string]healthz.Checker{
				"ok1": healthz.Ping,
				"bad1": func(req *http.Request) error {
					return errors.New("blech")
				},
			}}

			resp := requestTo(handler, "/")
			Expect(resp.Header().Get("Content-Type")).To(Equal(contentType))
			Expect(resp.Body.String()).To(Equal("[-]bad1 başarısız: sebep gizlendi\n[+]ok1 ok\nhealthz kontrolü başarısız\n"))
		})

		It("başka kontroller yoksa her zaman bir ping endpointi döndürmeli", func() {
			resp := requestTo(&healthz.Handler{}, "/?verbose=true")
			Expect(resp.Code).To(Equal(http.StatusOK))
			Expect(resp.Header().Get("Content-Type")).To(Equal(contentType))
			Expect(resp.Body.String()).To(Equal("[+]ping ok\nhealthz kontrolü geçti\n"))
		})
	})

	Describe("her kontrol için endpointler", func() {
		It("istenen kontrol sağlıklıysa ok dönmeli", func() {
			handler := &healthz.Handler{Checks: map[string]healthz.Checker{
				"okcheck": healthz.Ping,
			}}

			resp := requestTo(handler, "/okcheck")
			Expect(resp.Code).To(Equal(http.StatusOK))
			Expect(resp.Header().Get("Content-Type")).To(Equal(contentType))
			Expect(resp.Body.String()).To(Equal("ok"))
		})

		It("istenen kontrol sağlıksızsa hata dönmeli", func() {
			handler := &healthz.Handler{Checks: map[string]healthz.Checker{
				"failcheck": func(req *http.Request) error {
					return errors.New("blech")
				},
			}}

			resp := requestTo(handler, "/failcheck")
			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
			Expect(resp.Header().Get("Content-Type")).To(Equal(contentType))
			Expect(resp.Body.String()).To(Equal("iç sunucu hatası: blech\n"))
		})

		It("diğer kontrolleri dikkate almamalı", func() {
			handler := &healthz.Handler{Checks: map[string]healthz.Checker{
				"failcheck": func(req *http.Request) error {
					return errors.New("blech")
				},
				"okcheck": healthz.Ping,
			}}

			By("kötü endpointi kontrol edip başarısız olmasını beklemek")
			resp := requestTo(handler, "/failcheck")
			Expect(resp.Code).To(Equal(http.StatusInternalServerError))

			By("iyi endpointi kontrol edip başarılı olmasını beklemek")
			resp = requestTo(handler, "/okcheck")
			Expect(resp.Code).To(Equal(http.StatusOK))
		})

		It("bir kontrolle eşleşmeyen yollar için bulunamadı dönmeli", func() {
			handler := &healthz.Handler{}

			resp := requestTo(handler, "/doesnotexist")
			Expect(resp.Code).To(Equal(http.StatusNotFound))
		})

		It("başka kontroller yoksa her zaman bir ping endpointi döndürmeli", func() {
			resp := requestTo(&healthz.Handler{}, "/ping")
			Expect(resp.Code).To(Equal(http.StatusOK))
		})
	})
})
