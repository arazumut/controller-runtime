/*
2018 Kubernetes Yazarları Tarafından Telif Hakkı Saklıdır.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

    http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VERİLMEZ veya KOŞULSUZ OLARAK DAĞITILIR.
Lisans kapsamındaki izinler ve sınırlamalar hakkında daha fazla bilgi için Lisans'a bakınız.
*/

// Package controllerruntime, Kubernetes CRD'lerini ve birleştirilmiş/gömülü Kubernetes API'lerini
// manipüle eden Kubernetes tarzı denetleyiciler oluşturmak için araçlar sağlar.
//
// CRD'ler oluştururken yaygın kullanım durumları için kolay yardımcılar tanımlar,
// özelleştirilebilir soyutlama katmanlarının üzerine inşa edilmiştir. Yaygın durumlar kolay olmalı,
// ve yaygın olmayan durumlar mümkün olmalıdır. Genel olarak, controller-runtime kullanıcıları
// Kubernetes denetleyici en iyi uygulamalarına yönlendirmeye çalışır.
//
// # Başlarken
//
// controller-runtime için ana giriş noktası, denetleyiciler oluşturmaya başlamak için gereken
// tüm yaygın türleri içeren bu kök pakettir:
//
//	import (
//		ctrl "sigs.k8s.io/controller-runtime"
//	)
//
// Bu paketteki örnekler temel bir denetleyici kurulumunu anlatır.
// kubebuilder kitabı (https://book.kubebuilder.io) daha ayrıntılı yürüyüşler içerir.
//
// controller-runtime, yapıcılar yerine mantıklı varsayılanlara sahip yapıları tercih eder,
// bu nedenle controller-runtime'da doğrudan kullanılan yapıları görmek oldukça yaygındır.
//
// # Organizasyon
//
// Bu kitaplığın düzeninin kısa bir yürüyüşü aşağıda bulunabilir. Her paket, nasıl kullanılacağı
// hakkında daha fazla bilgi içerir.
//
// controller-runtime kullanımı ve denetleyici tasarımı hakkında sıkça sorulan sorular
// https://github.com/kubernetes-sigs/controller-runtime/blob/main/FAQ.md adresinde bulunabilir.
//
// # Yöneticiler
//
// Her denetleyici ve webhook nihayetinde bir Yönetici (pkg/manager) tarafından çalıştırılır.
// Bir yönetici, denetleyicileri ve webhooks'ları çalıştırmaktan ve paylaşılan önbellekler ve
// istemciler gibi ortak bağımlılıkları ayarlamaktan sorumludur, ayrıca lider seçimini yönetir
// (pkg/leaderelection). Yöneticiler genellikle bir sinyal işleyici bağlayarak pod sonlandırma
// sırasında denetleyicileri düzgün bir şekilde kapatacak şekilde yapılandırılır (pkg/manager/signals).
//
// # Denetleyiciler
//
// Denetleyiciler (pkg/controller), sonuçta yeniden uzlaştırma isteklerini tetiklemek için olayları
// (pkg/event) kullanır. Manuel olarak oluşturulabilirler, ancak genellikle olay kaynaklarını
// (pkg/source) olay işleyicilerine (pkg/handler) bağlamayı kolaylaştıran bir Yapıcı (pkg/builder)
// ile oluşturulurlar, örneğin "nesne sahibi için bir uzlaştırma isteği sıraya al". Predikatlar
// (pkg/predicate), hangi olayların gerçekten uzlaştırmaları tetikleyeceğini filtrelemek için
// kullanılabilir. Yaygın durumlar için önceden yazılmış yardımcılar ve gelişmiş durumlar için
// arayüzler ve yardımcılar vardır.
//
// # Uzlaştırıcılar
//
// Denetleyici mantığı, bir uzlaştırma İsteği içeren bir işlevi uygulayan Uzlaştırıcılar (pkg/reconcile)
// cinsinden uygulanır. Bir Uzlaştırıcı, uzlaştırılacak nesnenin adını ve ad alanını içeren bir
// uzlaştırma İsteği alır, nesneyi uzlaştırır ve bir Yanıt veya yeniden işleme için sıraya alınması
// gerekip gerekmediğini belirten bir hata döndürür.
//
// # İstemciler ve Önbellekler
//
// Uzlaştırıcılar, API nesnelerine erişmek için İstemciler (pkg/client) kullanır. Yönetici tarafından
// sağlanan varsayılan istemci, yerel paylaşılan bir önbellekten (pkg/cache) okur ve doğrudan API
// sunucusuna yazar, ancak yalnızca API sunucusuyla konuşan, önbelleği olmayan istemciler oluşturulabilir.
// Önbellek, izlenen nesnelerle otomatik olarak doldurulur ve diğer yapılandırılmış nesneler
// istendiğinde de doldurulur. Varsayılan bölünmüş istemci, yazma işlemleri sırasında önbelleği
// geçersiz kılmayı vaat etmez (ne de ardışık oluşturma/getirme tutarlılığı vaat eder) ve kod,
// bir oluşturma/güncelleme işlemini hemen takip eden bir get işleminin güncellenmiş kaynağı
// döndüreceğini varsaymamalıdır. Önbellekler ayrıca, yöneticiden elde edilen bir FieldIndexer
// (pkg/client) aracılığıyla oluşturulabilecek dizinlere sahip olabilir. Dizinler, belirli alanları
// ayarlanmış tüm nesneleri hızlı ve kolay bir şekilde aramak için kullanılabilir. Uzlaştırıcılar,
// yönetici kullanarak olay kaydedicilerini (pkg/recorder) alabilirler.
//
// # Şemalar
//
// İstemciler, Önbellekler ve Kubernetes'teki birçok şey, Go türlerini Kubernetes API Türleriyle
// (Grup-Sürüm-Türler, daha spesifik olarak) ilişkilendirmek için Şemalar (pkg/scheme) kullanır.
//
// # Webhooks
//
// Benzer şekilde, webhooks (pkg/webhook/admission) doğrudan uygulanabilir, ancak genellikle bir
// yapıcı (pkg/webhook/admission/builder) kullanılarak oluşturulur. Bir Yönetici tarafından
// yönetilen bir sunucu (pkg/webhook) aracılığıyla çalıştırılırlar.
//
// # Günlük Kaydı ve Metrikler
//
// controller-runtime'da günlük kaydı, yapılandırılmış günlükler aracılığıyla yapılır ve logr
// (https://pkg.go.dev/github.com/go-logr/logr) adlı bir dizi arayüz kullanır. controller-runtime,
// Zap'ı (https://go.uber.org/zap, pkg/log/zap) kullanmak için kolay kurulum sağlar, ancak
// controller-runtime için temel günlük kaydedici olarak herhangi bir logr uygulaması sağlayabilirsiniz.
//
// # Metrikler
//
// controller-runtime tarafından sağlanan Metrikler (pkg/metrics), controller-runtime'a özgü bir
// Prometheus metrik kayıt defterine kaydedilir. Yönetici, bunları bir HTTP uç noktası aracılığıyla
// sunabilir ve ek metrikler normal olarak bu Kayıt Defterine kaydedilebilir.
//
// # Test
//
// Denetleyicileriniz ve webhooks'larınız için kolayca entegrasyon ve birim testleri oluşturabilirsiniz
// test Ortamını (pkg/envtest) kullanarak. Bu, otomatik olarak bir etcd ve kube-apiserver kopyası
// kuracak ve API sunucusuna bağlanmak için doğru seçenekleri sağlayacaktır. Ginkgo test çerçevesiyle
// iyi çalışacak şekilde tasarlanmıştır, ancak herhangi bir test kurulumu ile çalışmalıdır.
package controllerruntime
