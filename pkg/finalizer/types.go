/*
2021 Kubernetes Yazarları tarafından oluşturulmuştur.
Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:
	http://www.apache.org/licenses/LICENSE-2.0
Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
bu yazılım Lisans kapsamında "OLDUĞU GİBİ" dağıtılmakta olup,
HERHANGİ BİR GARANTİ VERİLMEMEKTEDİR; ne açık ne de zımni.
Lisans kapsamındaki izinler ve sınırlamalar hakkında daha fazla bilgi için Lisansı inceleyin.
*/

package finalizer

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Registerer, bir anahtarın zaten kayıtlı olup olmadığını kontrol eden ve
// eğer kayıtlıysa hata veren; kayıtlı değilse, finalizer'ı sağlanan anahtar için
// finalizers haritasına değer olarak ekleyen Register'i tutar.
type Registerer interface {
	Register(key string, f Finalizer) error
}

// Finalizer, silme zaman damgasının ayarlanıp ayarlanmadığına bağlı olarak
// bir finalizer ekleyip/çıkartacak ve objenin güncellenmesi gerekip gerekmediğine
// dair bir gösterge döndürecek olan Finalize'i tutar.
type Finalizer interface {
	Finalize(context.Context, client.Object) (Result, error)
}

// Finalizers, sağlanan nesnenin bir silme zaman damgası olup olmadığını kontrol ederek
// tüm kayıtlı finalizer'ları finalize eden veya yoksa tüm kayıtlı finalizer'ları
// ayarlayan Registerer ve Finalizer'ı uygular.
type Finalizers interface {
	Registerer
	Finalizer
}
