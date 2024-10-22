/*
2021 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa uyarınca veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMADAN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
haklar ve sınırlamalar için Lisansa bakın.
*/

package authentication

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/go-logr/logr"
	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/klog/v2"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	errUnableToEncodeResponse = errors.New("yanıt kodlanamıyor")
)

// Request, bir kimlik doğrulama işleyicisi için girişi tanımlar.
// Sorgulanan nesneyi tanımlamak için bilgi içerir (grup, sürüm, tür, kaynak, alt kaynak, ad, ad alanı),
// ayrıca sorgulanan işlemi (örneğin Get, Create, vb.) ve nesnenin kendisini içerir.
type Request struct {
	authenticationv1.TokenReview
}

// Response, bir kimlik doğrulama işleyicisinin çıktısıdır.
// Belirli bir işlemin izin verilip verilmediğini belirten bir yanıt içerir.
type Response struct {
	authenticationv1.TokenReview
}

// Complete, TokenResponse'da henüz ayarlanmamış alanları doldurur. Yanıtı değiştirir.
func (r *Response) Complete(req Request) error {
	r.UID = req.UID

	return nil
}

// Handler, bir TokenReview işleyebilir.
type Handler interface {
	// Handle, bir TokenReview için bir yanıt verir.
	//
	// Sağlanan context, alınan http.Request'ten çıkarılır, bu da sarmalayıcı http.Handlers'ın
	// değerleri enjekte etmesine ve aşağı akış istek işlemenin iptalini kontrol etmesine olanak tanır.
	Handle(context.Context, Request) Response
}

// HandlerFunc, tek bir işlev kullanarak Handler arayüzünü uygular.
type HandlerFunc func(context.Context, Request) Response

var _ Handler = HandlerFunc(nil)

// Handle, TokenReview'u temel işlevi çağırarak işler.
func (f HandlerFunc) Handle(ctx context.Context, req Request) Response {
	return f(ctx, req)
}

// Webhook, her bir bireysel webhook'u temsil eder.
type Webhook struct {
	// Handler, bir kimlik doğrulama isteğini işleyerek kimlik doğrulandı mı yoksa doğrulanmadı mı olduğunu döndürür
	// ve potansiyel olarak işleyiciye uygulanacak yamaları içerir.
	Handler Handler

	// WithContextFunc, http.Request.Context()'i almanıza ve
	// ek bilgi eklemenize olanak tanır, böylece istek yolunu veya
	// başlıkları okuyabilirsiniz, bu da işleyici içinde onları okumanıza olanak tanır.
	WithContextFunc func(context.Context, *http.Request) context.Context

	setupLogOnce sync.Once
	log          logr.Logger
}

// Handle, TokenReview'u işler.
func (wh *Webhook) Handle(ctx context.Context, req Request) Response {
	resp := wh.Handler.Handle(ctx, req)
	if err := resp.Complete(req); err != nil {
		wh.getLogger(&req).Error(err, "yanıt kodlanamıyor")
		return Errored(errUnableToEncodeResponse)
	}

	return resp
}

// getLogger, enjekte edilen log ve LogConstructor'dan bir logger oluşturur.
func (wh *Webhook) getLogger(req *Request) logr.Logger {
	wh.setupLogOnce.Do(func() {
		if wh.log.GetSink() == nil {
			wh.log = logf.Log.WithName("authentication")
		}
	})

	return logConstructor(wh.log, req)
}

// logConstructor, verilen logger'a bazı yaygın olarak ilginç alanlar ekler.
func logConstructor(base logr.Logger, req *Request) logr.Logger {
	if req != nil {
		return base.WithValues("nesne", klog.KRef(req.Namespace, req.Name),
			"ad alanı", req.Namespace, "ad", req.Name,
			"kullanıcı", req.Status.User.Username,
			"istekID", req.UID,
		)
	}
	return base
}
