/*
2021 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisansa bakın.
*/

package main

import (
	"context"

	v1 "k8s.io/api/authentication/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/authentication"
)

// authenticator tokenreview'ları doğrular
type authenticator struct {
}

// NewAuthenticator yeni bir authenticator döner
// authenticator bir isteği token ile kabul eder.
func (a *authenticator) Handle(ctx context.Context, req authentication.Request) authentication.Response {
	if req.Spec.Token == "invalid" {
		return authentication.Unauthenticated("invalid geçersiz bir token", v1.UserInfo{})
	}
	return authentication.Authenticated("", v1.UserInfo{})
}
