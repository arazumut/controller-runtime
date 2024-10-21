/*
2018 Kubernetes Yazarları tarafından oluşturulmuştur.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

  http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ OLMAKSIZIN, açık veya zımni olarak.
Lisans kapsamındaki izinler ve kısıtlamalar için Lisansa bakınız.
*/

/*
Sahte paket, test için sahte bir istemci sağlar.

Sahte bir istemci, GroupVersionResource tarafından dizinlenmiş basit bir nesne deposu tarafından desteklenir.
İsteğe bağlı nesnelerle sahte bir istemci oluşturabilirsiniz.

	client := NewClientBuilder().WithScheme(scheme).WithObjects(initObjs...).Build()

İstemci arayüzünde tanımlanan yöntemleri çağırabilirsiniz.

Şüpheye düştüğünüzde, bu paketi kullanmamak ve bunun yerine
gerçek bir istemci ve API sunucusu ile envtest.Environment kullanmak neredeyse her zaman daha iyidir.

UYARI: ⚠️ Sahte İstemci ile Mevcut Sınırlamalar / Bilinen Sorunlar ⚠️
  - Bu istemcinin, işlenmiş ve işlenmemiş hataları test etmek için belirli hataları enjekte etmenin bir yolu yoktur.
  - Alt kaynaklar için bir miktar destek vardır, bu da aynı uzlaştırmada
    örneğin, meta verileri ve durumu güncellemeye çalışıyorsanız testlerde sorunlara neden olabilir.
  - Nesneleri oluştururken veya güncellerken herhangi bir OpenAPI doğrulaması yapılmaz.
  - ObjectMeta'nın `Generation` ve `ResourceVersion` düzgün çalışmaz, bu alanlara dayanan Yama veya Güncelleme
    işlemleri başarısız olur veya yanlış pozitif sonuçlar verir.
*/
package fake
