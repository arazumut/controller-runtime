/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package authentication

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	authenticationv1 "k8s.io/api/authentication/v1"
	authenticationv1beta1 "k8s.io/api/authentication/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var authenticationScheme = runtime.NewScheme()
var authenticationCodecs = serializer.NewCodecFactory(authenticationScheme)

// TokenReview kaynağı çoğunlukla birkaç KB boyutunda olması gereken bir taşıyıcı token içerir,
// bu yüzden bol miktarda tampon olması için 1 MB seçtik.
// Kullanım durumunuz daha büyük maksimum istek boyutları gerektiriyorsa,
// lütfen bir sorun açın (https://github.com/kubernetes-sigs/controller-runtime/issues/new).
const maxRequestSize = int64(1 * 1024 * 1024)

func init() {
	utilruntime.Must(authenticationv1.AddToScheme(authenticationScheme))
	utilruntime.Must(authenticationv1beta1.AddToScheme(authenticationScheme))
}

// Webhook, bir TokenReview isteğini işlemek için bir HTTP işleyicisidir.

var _ http.Handler = &Webhook{}

func (wh *Webhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if wh.WithContextFunc != nil {
		ctx = wh.WithContextFunc(ctx, r)
	}

	if r.Body == nil || r.Body == http.NoBody {
		err := errors.New("istek gövdesi boş")
		wh.getLogger(nil).Error(err, "kötü istek")
		wh.writeResponse(w, Errored(err))
		return
	}

	defer r.Body.Close()
	limitedReader := &io.LimitedReader{R: r.Body, N: maxRequestSize}
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		wh.getLogger(nil).Error(err, "gelen isteğin gövdesini okuyamıyor")
		wh.writeResponse(w, Errored(err))
		return
	}
	if limitedReader.N <= 0 {
		err := fmt.Errorf("istek varlığı çok büyük; limit %d bayt", maxRequestSize)
		wh.getLogger(nil).Error(err, "gelen isteğin gövdesini okuyamıyor; limit aşıldı")
		wh.writeResponse(w, Errored(err))
		return
	}

	// içerik türünün doğru olduğunu doğrula
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		err := fmt.Errorf("contentType=%s, beklenen application/json", contentType)
		wh.getLogger(nil).Error(err, "bilinmeyen içerik türü ile bir isteği işleyemiyor")
		wh.writeResponse(w, Errored(err))
		return
	}

	// Hem v1 hem de v1beta1 TokenReview türleri tamamen aynıdır, bu yüzden v1beta1 türü v1 türüne
	// kodlanabilir. v1beta1 API'si 1.19 itibariyle kullanımdan kaldırılmıştır ve authenticationv1.22'de
	// kaldırılacaktır. Ancak runtime codec'in decoder'i, bir Object'in TypeMeta'sı ayarlanmamışsa
	// tür adına göre hangi türe kodlanacağını tahmin eder. Kayıtlı olmayan bir türün TypeMeta'sını
	// v1 GVK'ya ayarlayarak, decoder bir v1beta1 TokenReview'u authenticationv1'e zorlayacaktır.
	// Gerçek TokenReview GVK, webhook yapılandırması birden fazla sürüme izin veriyorsa yazılı bir
	// yanıt vermek için kullanılacaktır, aksi takdirde bu yanıt başarısız olacaktır.
	req := Request{}
	ar := unversionedTokenReview{}
	// ekstra bir kopyadan kaçının
	ar.TokenReview = &req.TokenReview
	ar.SetGroupVersionKind(authenticationv1.SchemeGroupVersion.WithKind("TokenReview"))
	_, actualTokRevGVK, err := authenticationCodecs.UniversalDeserializer().Decode(body, nil, &ar)
	if err != nil {
		wh.getLogger(nil).Error(err, "isteği kodlayamıyor")
		wh.writeResponse(w, Errored(err))
		return
	}
	wh.getLogger(&req).V(5).Info("istek alındı")

	if req.Spec.Token == "" {
		err := errors.New("token boş")
		wh.getLogger(&req).Error(err, "kötü istek")
		wh.writeResponse(w, Errored(err))
		return
	}

	wh.writeResponseTyped(w, wh.Handle(ctx, req), actualTokRevGVK)
}

// writeResponse, yanıtı w'ye genel olarak yazar, yani GVK bilgilerini kodlamadan.
func (wh *Webhook) writeResponse(w io.Writer, response Response) {
	wh.writeTokenResponse(w, response.TokenReview)
}

// writeResponseTyped, yanıtı w'ye tokRevGVK'ya ayarlanmış GVK ile yazar, bu
// birden fazla TokenReview sürümüne izin veriliyorsa gereklidir.
func (wh *Webhook) writeResponseTyped(w io.Writer, response Response, tokRevGVK *schema.GroupVersionKind) {
	ar := response.TokenReview

	// Varsayılan olarak bir v1 TokenReview kullanın, aksi takdirde API sunucusu
	// webhook yapılandırması birden fazla TokenReview sürümüne izin veriyorsa isteği tanımayabilir.
	if tokRevGVK == nil || *tokRevGVK == (schema.GroupVersionKind{}) {
		ar.SetGroupVersionKind(authenticationv1.SchemeGroupVersion.WithKind("TokenReview"))
	} else {
		ar.SetGroupVersionKind(*tokRevGVK)
	}
	wh.writeTokenResponse(w, ar)
}

// writeTokenResponse, ar'yi w'ye yazar.
func (wh *Webhook) writeTokenResponse(w io.Writer, ar authenticationv1.TokenReview) {
	if err := json.NewEncoder(w).Encode(ar); err != nil {
		wh.getLogger(nil).Error(err, "yanıtı kodlayamıyor")
		wh.writeResponse(w, Errored(err))
	}
	res := ar
	wh.getLogger(nil).V(5).Info("yanıt yazıldı", "requestID", res.UID, "authenticated", res.Status.Authenticated)
}

// unversionedTokenReview, hem v1 hem de v1beta1 TokenReview türlerini kodlamak için kullanılır.
type unversionedTokenReview struct {
	*authenticationv1.TokenReview
}

var _ runtime.Object = &unversionedTokenReview{}
