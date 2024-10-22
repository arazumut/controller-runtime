/*
2018 Kubernetes Yazarları tarafından oluşturulmuştur.

Apache Lisansı, Sürüm 2.0 ("Lisans") altında lisanslanmıştır;
bu dosyayı ancak Lisans'a uygun olarak kullanabilirsiniz.
Lisans'ın bir kopyasını aşağıdaki adresten edinebilirsiniz:

    http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
herhangi bir garanti veya koşul olmaksızın, açık veya zımni.
Lisans kapsamındaki izinler ve sınırlamalar hakkında daha fazla bilgi için Lisans'a bakın.
*/

/*
pkg paketi, Denetleyiciler (Controllers) oluşturmak için kütüphaneler sağlar. Denetleyiciler, Kubernetes API'lerini uygular
ve Operatörler, İş Yükü API'leri, Yapılandırma API'leri, Otomatik Ölçekleyiciler ve daha fazlasını oluşturmak için temel oluşturur.

# İstemci (Client)

İstemci, Kubernetes nesnelerini okumak ve yazmak için bir Okuma + Yazma istemcisi sağlar.

# Önbellek (Cache)

Önbellek, yerel bir önbellekten nesneleri okumak için bir Okuma istemcisi sağlar.
Bir önbellek, önbelleği güncelleyen olaylara yanıt vermek için işleyiciler kaydedebilir.

# Yönetici (Manager)

Yönetici, bir Denetleyici oluşturmak için gereklidir ve Denetleyiciye istemciler, önbellekler, şemalar vb. gibi paylaşılan bağımlılıkları sağlar.
Denetleyiciler, Manager.Start çağrılarak Yönetici aracılığıyla başlatılmalıdır.

# Denetleyici (Controller)

Denetleyici, olaylara (nesne Oluşturma, Güncelleme, Silme) yanıt vererek ve nesnenin Spec'inde belirtilen durumun sistemin durumu ile eşleşmesini sağlayarak bir Kubernetes API'sini uygular. Bu işleme yeniden uzlaştırma (reconcile) denir.
Eğer eşleşmezlerse, Denetleyici, nesneleri eşleşmeleri için gerekli şekilde oluşturur/günceller/siler.

Denetleyiciler, yeniden uzlaştırma isteklerini (belirli bir nesnenin durumunu yeniden uzlaştırma istekleri) işleyen işçi kuyrukları olarak uygulanır.

HTTP işleyicilerinin aksine, Denetleyiciler olayları doğrudan işlemez, ancak nesneyi sonunda yeniden uzlaştırmak için istekleri sıraya alır.
Bu, birden fazla olayın bir araya getirilebileceği ve her yeniden uzlaştırma için sistemin tam durumunun okunması gerektiği anlamına gelir.

* Denetleyiciler, iş kuyruğundan çekilen işi gerçekleştirmek için bir Yeniden Uzlaştırıcı (Reconciler) gerektirir.

* Denetleyiciler, olaylara yanıt olarak yeniden uzlaştırma isteklerini sıraya almak için İzlemeler (Watches) yapılandırılmasını gerektirir.

# Webhook

Kabul Webhook'ları, Kubernetes API'lerini genişletmek için bir mekanizmadır. Webhook'lar, hedef olay türü (nesne Oluşturma, Güncelleme, Silme) ile yapılandırılabilir, API sunucusu belirli olaylar gerçekleştiğinde onlara Kabul İstekleri gönderir.
Webhook'lar, Kabul İnceleme isteklerinde gömülü nesneyi değiştirebilir ve (veya) doğrulayabilir ve yanıtı API sunucusuna geri gönderebilir.

İki tür kabul webhook'u vardır: değiştirme ve doğrulama kabul webhook'u.
Değiştirme webhook'u, API sunucusu tarafından kabul edilmeden önce bir çekirdek API nesnesini veya bir CRD örneğini değiştirmek için kullanılır.
Doğrulama webhook'u, bir nesnenin belirli gereksinimleri karşılayıp karşılamadığını doğrulamak için kullanılır.

* Kabul Webhook'ları, alınan Kabul İnceleme isteklerini işlemek için İşleyici(ler) gerektirir.

# Yeniden Uzlaştırıcı (Reconciler)

Yeniden Uzlaştırıcı, bir Denetleyiciye sağlanan ve herhangi bir zamanda bir nesnenin Adı ve Ad Alanı ile çağrılabilen bir işlevdir.
Çağrıldığında, Yeniden Uzlaştırıcı, sistemin durumunun, Yeniden Uzlaştırıcı çağrıldığında nesnede belirtilenle eşleşmesini sağlar.

Örnek: Bir ReplicaSet nesnesi için çağrılan Yeniden Uzlaştırıcı. ReplicaSet, 5 kopya belirtir ancak sistemde yalnızca 3 Pod vardır. Yeniden Uzlaştırıcı, 2 Pod daha oluşturur ve bunların Sahip Referansını, controller=true ile ReplicaSet'e işaret eder.

* Yeniden Uzlaştırıcı, bir Denetleyicinin tüm iş mantığını içerir.

* Yeniden Uzlaştırıcı tipik olarak tek bir nesne türü üzerinde çalışır. - örneğin, yalnızca ReplicaSet'leri yeniden uzlaştırır. Ayrı türler için ayrı Denetleyiciler kullanın. Başka nesnelerden yeniden uzlaştırma tetiklemek istiyorsanız, yeniden uzlaştırmayı tetikleyen nesneyi yeniden uzlaştırılan nesneye eşleyen bir eşleme (örneğin, sahip referansları) sağlayabilirsiniz.

* Yeniden Uzlaştırıcı, yeniden uzlaştırılacak nesnenin Adı/Ad Alanı sağlanır.

* Yeniden Uzlaştırıcı, yeniden uzlaştırmayı tetikleyen olay içeriği veya olay türü ile ilgilenmez.
- örneğin, bir ReplicaSet'in oluşturulup oluşturulmadığı veya güncellenip güncellenmediği önemli değildir, Yeniden Uzlaştırıcı her zaman sistemdeki Pod sayısını, çağrıldığı zamandaki nesnede belirtilenle karşılaştırır.

# Kaynak (Source)

resource.Source, olay akışı sağlayan bir Controller.Watch argümanıdır.
Olaylar tipik olarak Kubernetes API'lerini izlemekten gelir (örneğin, Pod Oluşturma, Güncelleme, Silme).

Örnek: source.Kind, bir GroupVersionKind için Kubernetes API İzleme uç noktasını kullanarak Oluşturma, Güncelleme, Silme olaylarını sağlar.

* Kaynak, Kubernetes nesneleri için tipik olarak İzleme API'si aracılığıyla bir olay akışı sağlar (örneğin, nesne Oluşturma, Güncelleme, Silme).

* Kullanıcılar, neredeyse tüm durumlar için kendi uygulamalarını yapmak yerine sağlanan Kaynak uygulamalarını kullanmalıdır.

# Olay İşleyici (EventHandler)

handler.EventHandler, olaylara yanıt olarak yeniden uzlaştırma isteklerini sıraya alan bir Controller.Watch argümanıdır.

Örnek: bir Kaynaktan gelen bir Pod Oluşturma olayı, ad/Ad Alanı içeren bir yeniden uzlaştırma isteğini sıraya alan eventhandler.EnqueueHandler'a sağlanır.

* Olay İşleyiciler, olayları bir veya daha fazla nesne için yeniden uzlaştırma isteklerini sıraya alarak işler.

* Olay İşleyiciler, bir nesne için bir olayı aynı türdeki bir nesne için bir yeniden uzlaştırma isteğine eşleyebilir.

* Olay İşleyiciler, bir nesne için bir olayı farklı türdeki bir nesne için bir yeniden uzlaştırma isteğine eşleyebilir - örneğin, bir Pod olayını sahip ReplicaSet için bir yeniden uzlaştırma isteğine eşleyebilir.

* Olay İşleyiciler, bir nesne için bir olayı aynı veya farklı türdeki birden fazla nesne için yeniden uzlaştırma isteklerine eşleyebilir - örneğin, bir Node olayını küme yeniden boyutlandırma olaylarına yanıt veren nesnelere eşleyebilir.

* Kullanıcılar, neredeyse tüm durumlar için kendi uygulamalarını yapmak yerine sağlanan Olay İşleyici uygulamalarını kullanmalıdır.

# Öngörü (Predicate)

predicate.Predicate, olayları filtreleyen isteğe bağlı bir Controller.Watch argümanıdır. Bu, yaygın filtrelerin yeniden kullanılmasını ve birleştirilmesini sağlar.

* Öngörü, bir olayı alır ve bir bool (sıraya almak için true) döndürür.

* Öngörüler isteğe bağlı argümanlardır.

* Kullanıcılar, sağlanan Öngörü uygulamalarını kullanmalıdır, ancak ek Öngörüler uygulayabilirler, örneğin nesil değişti, etiket seçiciler değişti vb.

# PodController Diyagramı

Kaynak olay sağlar:

* &source.KindSource{&v1.Pod{}} -> (Pod foo/bar Oluşturma Olayı)

Olay İşleyici İsteği sıraya alır:

* &handler.EnqueueRequestForObject{} -> (reconcile.Request{types.NamespaceName{Name: "foo", Namespace: "bar"}})

Yeniden Uzlaştırıcı, İstek ile çağrılır:

* Reconciler(reconcile.Request{types.NamespaceName{Name: "foo", Namespace: "bar"}})

# Kullanım

Aşağıdaki örnek, Pod veya ReplicaSet olaylarına yanıt olarak ReplicaSet nesnelerini yeniden uzlaştıran yeni bir Denetleyici programı oluşturmayı gösterir. Yeniden Uzlaştırıcı işlevi, ReplicaSet'e bir etiket ekler.

Kullanım örneği için examples/builtins/main.go dosyasına bakın.

Denetleyici Örneği:

1. ReplicaSet ve Pod Kaynaklarını İzleyin

1.1 ReplicaSet -> handler.EnqueueRequestForObject - ReplicaSet Ad Alanı ve Adı ile bir İstek sıraya alın.

1.2 Pod (ReplicaSet tarafından oluşturulan) -> handler.EnqueueRequestForOwnerHandler - Sahip ReplicaSet Ad Alanı ve Adı ile bir İstek sıraya alın.

2. Bir olaya yanıt olarak ReplicaSet'i yeniden uzlaştırın

2.1 ReplicaSet nesnesi oluşturuldu -> ReplicaSet'i okuyun, Pod'ları okumaya çalışın -> eksikse Pod'ları oluşturun.

2.2 Pod'ların oluşturulmasıyla tetiklenen Yeniden Uzlaştırıcı -> ReplicaSet ve Pod'ları okuyun, hiçbir şey yapmayın.

2.3 Başka bir aktör tarafından Pod'ların silinmesiyle tetiklenen Yeniden Uzlaştırıcı -> ReplicaSet ve Pod'ları okuyun, yedek Pod'ları oluşturun.

# İzleme ve Olay İşleme

Denetleyiciler, birden fazla türde nesneyi izleyebilir (örneğin, Pod'lar, ReplicaSet'ler ve Dağıtımlar), ancak yalnızca tek bir Türü yeniden uzlaştırır. Bir Türdeki nesne, başka bir Türdeki nesnelerdeki değişikliklere yanıt olarak güncellenmesi gerektiğinde, bir EnqueueRequestsFromMapFunc, olayları bir türden diğerine eşlemek için kullanılabilir. örneğin, bir küme yeniden boyutlandırma olayına (Node ekleme/silme) yanıt olarak bazı API örneklerinin tümünü yeniden uzlaştırmak.

Bir Dağıtım Denetleyicisi, bir EnqueueRequestForObject ve EnqueueRequestForOwner kullanabilir:

* Dağıtım Olaylarını İzleyin - Dağıtımın Ad Alanı ve Adını sıraya alın.

* ReplicaSet Olaylarını İzleyin - ReplicaSet'i oluşturan Dağıtımın Ad Alanı ve Adını sıraya alın (örneğin, Sahip).

Not: yeniden uzlaştırma istekleri sıraya alındığında yinelenir. Aynı ReplicaSet için birçok Pod Olayı, yalnızca 1 yeniden uzlaştırma çağrısını tetikleyebilir, çünkü her Olay, aynı ReplicaSet için yeniden uzlaştırma isteğini sıraya almaya çalışır.

# Denetleyici Yazma İpuçları

Yeniden Uzlaştırıcı Çalışma Zamanı Karmaşıklığı:

* Bir O(1) yeniden uzlaştırmayı N kez (örneğin, N farklı nesne üzerinde) gerçekleştiren Denetleyiciler yazmak, bir O(N) yeniden uzlaştırmayı 1 kez (örneğin, N diğer nesneyi yöneten tek bir nesne üzerinde) gerçekleştirmekten daha iyidir.

* Örnek: Bir Node eklendiğinde tüm Hizmetleri güncellemeniz gerekiyorsa - Hizmetleri yeniden uzlaştırın ancak Node'ları İzleyin (Hizmet nesne adı/Ad Alanlarına dönüştürülmüş) yerine Node'ları yeniden uzlaştırın ve Hizmetleri güncelleyin.

Olay Çoklama:

* Aynı Ad/Ad Alanı için yeniden uzlaştırma istekleri sıraya alındığında birleştirilir ve yinelenir. Bu, Denetleyicilerin tek bir nesne için yüksek hacimli olayları zarif bir şekilde işlemesini sağlar. Birden fazla olay Kaynağını tek bir nesne Türüne çoklamak, farklı nesne türlerinden gelen olaylar için istekleri birleştirir.

* Örnek: Bir ReplicaSet için Pod olayları, bir ReplicaSet Adı/Ad Alanına dönüştürülür, böylece ReplicaSet, birden fazla Pod'dan gelen birden fazla olay için yalnızca 1 kez yeniden uzlaştırılır.
*/
package pkg
