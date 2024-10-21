/*
2018 Kubernetes Yazarları tarafından oluşturulmuştur.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

    http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni olarak.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisansa bakınız.
*/

/*
handler paketi, Kubernetes API'lerini izleyerek gözlemlenen Oluşturma, Güncelleme, Silme Olaylarına yanıt olarak reconcile.Request'leri sıraya alan EventHandler'ları tanımlar.
Kullanıcılar, reconcile.Request iş öğeleri oluşturmak ve sıraya almak için Controller.Watch'a bir source.Source ve handler.EventHandler sağlamalıdır.

Genel olarak, çoğu kullanım durumu için aşağıdaki hazır olay işleyiciler yeterli olacaktır:

EventHandlers:

EnqueueRequestForObject - Olaydaki nesnenin Adını ve Ad Alanını içeren bir reconcile.Request'i sıraya alır. Bu, Olayın kaynağı olan nesnenin (örneğin, oluşturulan / silinen / güncellenen nesne) yeniden uzlaştırılmasına neden olacaktır.

EnqueueRequestForOwner - Olaydaki nesnenin Sahibinin Adını ve Ad Alanını içeren bir reconcile.Request'i sıraya alır. Bu, Olayın kaynağı olan nesnenin sahibinin (örneğin, nesneyi oluşturan sahip nesne) yeniden uzlaştırılmasına neden olacaktır.

EnqueueRequestsFromMapFunc - Olaydaki nesneye karşı çalıştırılan kullanıcı tarafından sağlanan bir dönüşüm fonksiyonundan kaynaklanan reconcile.Request'leri sıraya alır. Bu, kaynak nesnenin bir dönüşümünden tanımlanan rastgele bir nesne koleksiyonunun yeniden uzlaştırılmasına neden olacaktır.
*/
package handler
