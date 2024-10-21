/*
2019 Kubernetes Yazarları tarafından oluşturulmuştur.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil altındaki haklar ve
sınırlamalar için Lisansa bakınız.
*/

/*
Dönüşüm paketi, bir API Türünün desteklenmesi için uygulaması gereken
arayüz tanımlarını sağlar. Bu, pkg/webhook/conversion altında tanımlanan
genel dönüşüm webhook işleyicisi tarafından desteklenir.
*/
package conversion

import "k8s.io/apimachinery/pkg/runtime"

// Convertible, bir türün dönüştürülebilir olma yeteneğini tanımlar, yani bir hub türüne dönüştürülebilir.
type Convertible interface {
	runtime.Object
	ConvertTo(dst Hub) error
	ConvertFrom(src Hub) error
}

// Hub, belirli bir türün dönüşüm için hub türü olduğunu belirtir. Bu, tüm dönüşümlerin
// önce hub türüne dönüştürüleceği, ardından hedef türe dönüştürüleceği anlamına gelir.
// Hub türü dışındaki tüm türler Convertible arayüzünü uygulamalıdır.
type Hub interface {
	runtime.Object
	Hub()
}
