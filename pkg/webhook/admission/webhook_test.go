/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa veya yazılı izin gereği olmadıkça,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMADAN.
Lisans kapsamındaki belirli dil izinlerini ve
sınırlamaları görmek için Lisansı okuyun.
*/

package admission

import (
	"context"
	"io"
	"net/http"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"gomodules.xyz/jsonpatch/v2"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	machinerytypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var _ = Describe("Admission Webhooks", func() {
	var (
		logBuffer  *gbytes.Buffer
		testLogger logr.Logger
	)

	BeforeEach(func() {
		logBuffer = gbytes.NewBuffer()
		testLogger = zap.New(zap.JSONEncoder(), zap.WriteTo(io.MultiWriter(logBuffer, GinkgoWriter)))
	})

	allowHandler := func() *Webhook {
		handler := &fakeHandler{
			fn: func(ctx context.Context, req Request) Response {
				return Response{
					AdmissionResponse: admissionv1.AdmissionResponse{
						Allowed: true,
					},
				}
			},
		}
		webhook := &Webhook{
			Handler: handler,
		}

		return webhook
	}

	It("handler'ı çağırarak bir yanıt almalı", func() {
		By("izin veren bir handler ile bir webhook kurmak")
		webhook := allowHandler()

		By("webhook'u çağırmak")
		resp := webhook.Handle(context.Background(), Request{})

		By("isteği kabul ettiğini kontrol etmek")
		Expect(resp.Allowed).To(BeTrue())
	})

	It("yanıtın UID'sinin isteğin UID'sine ayarlandığından emin olmalı", func() {
		By("bir webhook kurmak")
		webhook := allowHandler()

		By("webhook'u çağırmak")
		resp := webhook.Handle(context.Background(), Request{AdmissionRequest: admissionv1.AdmissionRequest{UID: "foobar"}})

		By("yanıtın isteğin UID'sini paylaştığını kontrol etmek")
		Expect(resp.UID).To(Equal(machinerytypes.UID("foobar")))
	})

	It("bir yanıtın durumu sağlanmadıysa durumu doldurmalı", func() {
		By("bir webhook kurmak")
		webhook := allowHandler()

		By("webhook'u çağırmak")
		resp := webhook.Handle(context.Background(), Request{})

		By("yanıtın isteğin UID'sini paylaştığını kontrol etmek")
		Expect(resp.Result).To(Equal(&metav1.Status{Code: http.StatusOK}))
	})

	It("bir yanıtın durumunu geçersiz kılmamalı", func() {
		By("durum belirten bir webhook kurmak")
		webhook := &Webhook{
			Handler: HandlerFunc(func(ctx context.Context, req Request) Response {
				return Response{
					AdmissionResponse: admissionv1.AdmissionResponse{
						Allowed: true,
						Result:  &metav1.Status{Message: "Ground Control to Major Tom"},
					},
				}
			}),
		}

		By("webhook'u çağırmak")
		resp := webhook.Handle(context.Background(), Request{})

		By("mesajın bozulmadığını kontrol etmek")
		Expect(resp.Result).NotTo(BeNil())
		Expect(resp.Result.Message).To(Equal("Ground Control to Major Tom"))
	})

	It("yama işlemlerini tek bir jsonpatch blobuna seri hale getirmeli", func() {
		By("yama yapan bir handler ile bir webhook kurmak")
		webhook := &Webhook{
			Handler: HandlerFunc(func(ctx context.Context, req Request) Response {
				return Patched("", jsonpatch.Operation{Operation: "add", Path: "/a", Value: 2}, jsonpatch.Operation{Operation: "replace", Path: "/b", Value: 4})
			}),
		}

		By("webhook'u çağırmak")
		resp := webhook.Handle(context.Background(), Request{})

		By("yanıtta bir JSON yamasının doldurulduğunu kontrol etmek")
		patchType := admissionv1.PatchTypeJSONPatch
		Expect(resp.PatchType).To(Equal(&patchType))
		Expect(resp.Patch).To(Equal([]byte(`[{"op":"add","path":"/a","value":2},{"op":"replace","path":"/b","value":4}]`)))
	})

	It("istek logger'ını bağlam üzerinden geçirmeli", func() {
		By("istek logger'ını kullanan bir webhook kurmak")
		webhook := &Webhook{
			Handler: HandlerFunc(func(ctx context.Context, req Request) Response {
				logf.FromContext(ctx).Info("İstek alındı")

				return Response{
					AdmissionResponse: admissionv1.AdmissionResponse{
						Allowed: true,
					},
				}
			}),
			log: testLogger,
		}

		By("webhook'u çağırmak")
		resp := webhook.Handle(context.Background(), Request{AdmissionRequest: admissionv1.AdmissionRequest{
			UID:       "test123",
			Name:      "foo",
			Namespace: "bar",
			Resource: metav1.GroupVersionResource{
				Group:    "apps",
				Version:  "v1",
				Resource: "deployments",
			},
			UserInfo: authenticationv1.UserInfo{
				Username: "tim",
			},
		}})
		Expect(resp.Allowed).To(BeTrue())

		By("log mesajının istek alanlarını içerdiğini kontrol etmek")
		Eventually(logBuffer).Should(gbytes.Say(`"msg":"İstek alındı","object":{"name":"foo","namespace":"bar"},"namespace":"bar","name":"foo","resource":{"group":"apps","version":"v1","resource":"deployments"},"user":"tim","requestID":"test123"}`))
	})

	It("LogConstructor tarafından oluşturulan istek logger'ını bağlam üzerinden geçirmeli", func() {
		By("istek logger'ını kullanan bir webhook kurmak")
		webhook := &Webhook{
			Handler: HandlerFunc(func(ctx context.Context, req Request) Response {
				logf.FromContext(ctx).Info("İstek alındı")

				return Response{
					AdmissionResponse: admissionv1.AdmissionResponse{
						Allowed: true,
					},
				}
			}),
			LogConstructor: func(base logr.Logger, req *Request) logr.Logger {
				return base.WithValues("operation", req.Operation, "requestID", req.UID)
			},
			log: testLogger,
		}

		By("webhook'u çağırmak")
		resp := webhook.Handle(context.Background(), Request{AdmissionRequest: admissionv1.AdmissionRequest{
			UID:       "test123",
			Operation: admissionv1.Create,
		}})
		Expect(resp.Allowed).To(BeTrue())

		By("log mesajının istek alanlarını içerdiğini kontrol etmek")
		Eventually(logBuffer).Should(gbytes.Say(`"msg":"İstek alındı","operation":"CREATE","requestID":"test123"`))
	})

	Describe("panik kurtarma", func() {
		It("RecoverPanic varsayılan olarak true olduğunda panikten kurtulmalı", func() {
			panicHandler := func() *Webhook {
				handler := &fakeHandler{
					fn: func(ctx context.Context, req Request) Response {
						panic("sahte panik testi")
					},
				}
				webhook := &Webhook{
					Handler: handler,
					// RecoverPanic varsayılan olarak true.
				}

				return webhook
			}

			By("panikleyen bir handler ile bir webhook kurmak")
			webhook := panicHandler()

			By("webhook'u çağırmak")
			resp := webhook.Handle(context.Background(), Request{})

			By("isteğin hata verdiğini kontrol etmek")
			Expect(resp.Allowed).To(BeFalse())
			Expect(resp.Result.Code).To(Equal(int32(http.StatusInternalServerError)))
			Expect(resp.Result.Message).To(Equal("panik: sahte panik testi [kurtarıldı]"))
		})

		It("RecoverPanic true olduğunda panikten kurtulmalı", func() {
			panicHandler := func() *Webhook {
				handler := &fakeHandler{
					fn: func(ctx context.Context, req Request) Response {
						panic("sahte panik testi")
					},
				}
				webhook := &Webhook{
					Handler:      handler,
					RecoverPanic: ptr.To[bool](true),
				}

				return webhook
			}

			By("panikleyen bir handler ile bir webhook kurmak")
			webhook := panicHandler()

			By("webhook'u çağırmak")
			resp := webhook.Handle(context.Background(), Request{})

			By("isteğin hata verdiğini kontrol etmek")
			Expect(resp.Allowed).To(BeFalse())
			Expect(resp.Result.Code).To(Equal(int32(http.StatusInternalServerError)))
			Expect(resp.Result.Message).To(Equal("panik: sahte panik testi [kurtarıldı]"))
		})

		It("RecoverPanic false olduğunda panikten kurtulmamalı", func() {
			panicHandler := func() *Webhook {
				handler := &fakeHandler{
					fn: func(ctx context.Context, req Request) Response {
						panic("sahte panik testi")
					},
				}
				webhook := &Webhook{
					Handler:      handler,
					RecoverPanic: ptr.To[bool](false),
				}

				return webhook
			}

			By("panikleyen bir handler ile bir webhook kurmak")
			defer func() {
				Expect(recover()).ShouldNot(BeNil())
			}()
			webhook := panicHandler()

			By("webhook'u çağırmak")
			webhook.Handle(context.Background(), Request{})
		})
	})
})

var _ = Describe("admission.Request'i bağlama yazabilmeli/okuyabilmeli", func() {
	ctx := context.Background()
	testRequest := Request{
		admissionv1.AdmissionRequest{
			UID: "test-uid",
		},
	}

	ctx = NewContextWithRequest(ctx, testRequest)

	gotRequest, err := RequestFromContext(ctx)
	Expect(err).To(Not(HaveOccurred()))
	Expect(gotRequest).To(Equal(testRequest))
})
