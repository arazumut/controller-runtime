/*
2021 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") altında lisanslanmıştır;
bu dosyayı ancak Lisansa uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans altında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisansa bakınız.
*/

package authentication

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	authenticationv1 "k8s.io/api/authentication/v1"
)

var _ = Describe("Kimlik Doğrulama Webhook'ları", func() {

	const (
		gvkJSONv1 = `"kind":"TokenReview","apiVersion":"authentication.k8s.io/v1"`
	)

	Describe("HTTP Handler", func() {
		var respRecorder *httptest.ResponseRecorder
		webhook := &Webhook{
			Handler: nil,
		}
		BeforeEach(func() {
			respRecorder = &httptest.ResponseRecorder{
				Body: bytes.NewBuffer(nil),
			}
		})

		It("boş bir gövde verildiğinde kötü istek döndürmelidir", func() {
			req := &http.Request{Body: nil}

			expected := `{"metadata":{"creationTimestamp":null},"spec":{},"status":{"user":{},"error":"request body is empty"}}
`
			webhook.ServeHTTP(respRecorder, req)
			Expect(respRecorder.Body.String()).To(Equal(expected))
		})

		It("yanlış içerik türü verildiğinde kötü istek döndürmelidir", func() {
			req := &http.Request{
				Header: http.Header{"Content-Type": []string{"application/foo"}},
				Method: http.MethodPost,
				Body:   nopCloser{Reader: bytes.NewBuffer(nil)},
			}

			expected := `{"metadata":{"creationTimestamp":null},"spec":{},"status":{"user":{},"error":"contentType=application/foo, expected application/json"}}
`
			webhook.ServeHTTP(respRecorder, req)
			Expect(respRecorder.Body.String()).To(Equal(expected))
		})

		It("çözülemeyen bir gövde verildiğinde kötü istek döndürmelidir", func() {
			req := &http.Request{
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Method: http.MethodPost,
				Body:   nopCloser{Reader: bytes.NewBufferString("{")},
			}

			expected := `{"metadata":{"creationTimestamp":null},"spec":{},"status":{"user":{},"error":"couldn't get version/kind; json parse error: unexpected end of JSON input"}}
`
			webhook.ServeHTTP(respRecorder, req)
			Expect(respRecorder.Body.String()).To(Equal(expected))
		})

		It("boş bir token verildiğinde kötü istek döndürmelidir", func() { // Bu test ismi düzeltilmiştir
			req := &http.Request{
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Method: http.MethodPost,
				Body:   nopCloser{Reader: bytes.NewBufferString(`{"spec":{"token":""}}`)},
			}

			expected := `{"metadata":{"creationTimestamp":null},"spec":{},"status":{"user":{},"error":"token is empty"}}
`
			webhook.ServeHTTP(respRecorder, req)
			Expect(respRecorder.Body.String()).To(Equal(expected))
		})

		It("NoBody verildiğinde hata döndürmelidir", func() {
			req := &http.Request{
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Method: http.MethodPost,
				Body:   http.NoBody,
			}

			expected := `{"metadata":{"creationTimestamp":null},"spec":{},"status":{"user":{},"error":"request body is empty"}}
`
			webhook.ServeHTTP(respRecorder, req)
			Expect(respRecorder.Body.String()).To(Equal(expected))
		})

		It("sonsuz bir gövde verildiğinde hata döndürmelidir", func() {
			req := &http.Request{
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Method: http.MethodPost,
				Body:   nopCloser{Reader: rand.Reader},
			}

			expected := `{"metadata":{"creationTimestamp":null},"spec":{},"status":{"user":{},"error":"request entity is too large; limit is 1048576 bytes"}}
`
			webhook.ServeHTTP(respRecorder, req)
			Expect(respRecorder.Body.String()).To(Equal(expected))
		})

		It("handler tarafından verilen yanıtı v1 sürümüne varsayılan olarak döndürmelidir", func() {
			req := &http.Request{
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Method: http.MethodPost,
				Body:   nopCloser{Reader: bytes.NewBufferString(`{"spec":{"token":"foobar"}}`)},
			}
			webhook := &Webhook{
				Handler: &fakeHandler{},
			}

			expected := fmt.Sprintf(`{%s,"metadata":{"creationTimestamp":null},"spec":{},"status":{"authenticated":true,"user":{}}}
`, gvkJSONv1)

			webhook.ServeHTTP(respRecorder, req)
			Expect(respRecorder.Body.String()).To(Equal(expected))
		})

		It("handler tarafından verilen v1 yanıtını döndürmelidir", func() {
			req := &http.Request{
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Method: http.MethodPost,
				Body:   nopCloser{Reader: bytes.NewBufferString(fmt.Sprintf(`{%s,"spec":{"token":"foobar"}}`, gvkJSONv1))},
			}
			webhook := &Webhook{
				Handler: &fakeHandler{},
			}

			expected := fmt.Sprintf(`{%s,"metadata":{"creationTimestamp":null},"spec":{},"status":{"authenticated":true,"user":{}}}
`, gvkJSONv1)
			webhook.ServeHTTP(respRecorder, req)
			Expect(respRecorder.Body.String()).To(Equal(expected))
		})

		It("HTTP isteğinden gelen Context'i sunmalıdır, varsa", func() {
			req := &http.Request{
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Method: http.MethodPost,
				Body:   nopCloser{Reader: bytes.NewBufferString(`{"spec":{"token":"foobar"}}`)},
			}
			type ctxkey int
			const key ctxkey = 1
			const value = "from-ctx"
			webhook := &Webhook{
				Handler: &fakeHandler{
					fn: func(ctx context.Context, req Request) Response {
						<-ctx.Done()
						return Authenticated(ctx.Value(key).(string), authenticationv1.UserInfo{})
					},
				},
			}

			expected := fmt.Sprintf(`{%s,"metadata":{"creationTimestamp":null},"spec":{},"status":{"authenticated":true,"user":{},"error":%q}}
`, gvkJSONv1, value)

			ctx, cancel := context.WithCancel(context.WithValue(context.Background(), key, value))
			cancel()
			webhook.ServeHTTP(respRecorder, req.WithContext(ctx))
			Expect(respRecorder.Body.String()).To(Equal(expected))
		})

		It("HTTP isteğinden gelen Context'i değiştirmelidir, eğer fonksiyon sağlanmışsa", func() {
			req := &http.Request{
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Method: http.MethodPost,
				Body:   nopCloser{Reader: bytes.NewBufferString(`{"spec":{"token":"foobar"}}`)},
			}
			type ctxkey int
			const key ctxkey = 1
			webhook := &Webhook{
				Handler: &fakeHandler{
					fn: func(ctx context.Context, req Request) Response {
						return Authenticated(ctx.Value(key).(string), authenticationv1.UserInfo{})
					},
				},
				WithContextFunc: func(ctx context.Context, r *http.Request) context.Context {
					return context.WithValue(ctx, key, r.Header["Content-Type"][0])
				},
			}

			expected := fmt.Sprintf(`{%s,"metadata":{"creationTimestamp":null},"spec":{},"status":{"authenticated":true,"user":{},"error":%q}}
`, gvkJSONv1, "application/json")

			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			webhook.ServeHTTP(respRecorder, req.WithContext(ctx))
			Expect(respRecorder.Body.String()).To(Equal(expected))
		})
	})
})

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

type fakeHandler struct {
	invoked bool
	fn      func(context.Context, Request) Response
}

func (h *fakeHandler) Handle(ctx context.Context, req Request) Response {
	h.invoked = true
	if h.fn != nil {
		return h.fn(ctx, req)
	}
	return Response{TokenReview: authenticationv1.TokenReview{
		Status: authenticationv1.TokenReviewStatus{
			Authenticated: true,
		},
	}}
}
