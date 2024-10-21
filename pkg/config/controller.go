/*
2023 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VERİLMEKSİZİN; ne açık ne de zımni.
Lisans kapsamındaki izin ve sınırlamalarla ilgili daha fazla bilgi için Lisansa bakınız.
*/

package config

import "time"

// Controller, bir kontrolcü için yapılandırma seçeneklerini içerir.
type Controller struct {
	// SkipNameValidation, her kontrolcü isminin benzersiz olmasını sağlayan isim doğrulamasını atlamaya izin verir.
	// Benzersiz kontrolcü isimleri, bir kontrolcü için benzersiz metrikler ve günlükler almak için önemlidir.
	// Kontrolcü üzerindeki SkipNameValidation ayarı ile geçersiz kılınabilir.
	// Kontrolcü ve Yönetici üzerindeki SkipNameValidation ayarları belirlenmemişse varsayılan olarak false olur.
	SkipNameValidation *bool

	// GroupKindConcurrency, bir türden o kontrolcü için izin verilen eşzamanlı uzlaştırma sayısına bir haritadır.
	//
	// Bir kontrolcü, bu yönetici içinde oluşturucu yardımcıları kullanılarak kaydedildiğinde,
	// kullanıcılar For(...) çağrısında kontrolcünün uzlaştırdığı türü belirtmelidir.
	// Geçilen nesnenin türü bu haritadaki anahtarlardan biriyle eşleşirse, o kontrolcü için eşzamanlılık belirtilen sayıya ayarlanır.
	//
	// Anahtarın, GroupKind.String() ile tutarlı bir biçimde olması beklenir,
	// örneğin, uygulamalar grubundaki ReplicaSet (sürümden bağımsız olarak) `ReplicaSet.apps` olacaktır.
	GroupKindConcurrency map[string]int

	// MaxConcurrentReconciles, çalıştırılabilecek maksimum eşzamanlı uzlaştırma sayısıdır. Varsayılan olarak 1'dir.
	MaxConcurrentReconciles int

	// CacheSyncTimeout, önbelleklerin senkronizasyonunu beklemek için belirlenen zaman sınırını ifade eder.
	// Ayarlanmazsa varsayılan olarak 2 dakika olur.
	CacheSyncTimeout time.Duration

	// RecoverPanic, uzlaştırma tarafından neden olunan paniklerin kurtarılıp kurtarılmayacağını belirtir.
	// Kontrolcü üzerindeki RecoverPanic ayarı ile geçersiz kılınabilir.
	// Kontrolcü ve Yönetici üzerindeki RecoverPanic ayarları belirlenmemişse varsayılan olarak true olur.
	RecoverPanic *bool

	// NeedLeaderElection, kontrolcünün lider seçimini kullanması gerekip gerekmediğini belirtir.
	// Varsayılan olarak true'dur, bu da kontrolcünün lider seçimini kullanacağı anlamına gelir.
	NeedLeaderElection *bool
}
