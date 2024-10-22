/*
2021 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izinle gerekli olmadıkça,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMADAN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
haklar ve sınırlamalar için Lisansa bakın.
*/

package authentication

import (
	authenticationv1 "k8s.io/api/authentication/v1"
)

// Authenticated, verilen tokenın geçerli olduğunu belirten bir yanıt oluşturur.
func Authenticated(reason string, user authenticationv1.UserInfo) Response {
	return ReviewResponse(true, user, reason)
}

// Unauthenticated, verilen tokenın geçerli olmadığını belirten bir yanıt oluşturur.
func Unauthenticated(reason string, user authenticationv1.UserInfo) Response {
	return ReviewResponse(false, authenticationv1.UserInfo{}, reason)
}

// Errored, bir isteği hata işleme için yeni bir Yanıt oluşturur.
func Errored(err error) Response {
	return Response{
		TokenReview: authenticationv1.TokenReview{
			Spec: authenticationv1.TokenReviewSpec{},
			Status: authenticationv1.TokenReviewStatus{
				Authenticated: false,
				Error:         err.Error(),
			},
		},
	}
}

// ReviewResponse, bir isteği kabul etmek için bir yanıt döndürür.
func ReviewResponse(authenticated bool, user authenticationv1.UserInfo, err string, audiences ...string) Response {
	resp := Response{
		TokenReview: authenticationv1.TokenReview{
			Status: authenticationv1.TokenReviewStatus{
				Authenticated: authenticated,
				User:          user,
				Audiences:     audiences,
			},
		},
	}
	if len(err) > 0 {
		resp.TokenReview.Status.Error = err
	}
	return resp
}
