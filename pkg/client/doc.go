/*
2018 Kubernetes Yazarları tarafından oluşturulmuştur.

Apache Lisansı, Sürüm 2.0 ("Lisans") kapsamında lisanslanmıştır;
bu dosyayı ancak Lisans'a uygun şekilde kullanabilirsiniz.
Lisans'ın bir kopyasını aşağıdaki adreste bulabilirsiniz:

    http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VERİLMEZ; açık veya zımni garantiler dahil.
Lisans kapsamındaki izin ve sınırlamalar için Lisans'a bakınız.
*/

// Paket client, Kubernetes API sunucuları ile etkileşim için işlevsellik içerir.
//
// # İstemciler
//
// İstemciler iki arayüze ayrılmıştır -- Okuyucular ve Yazıcılar. Okuyucular
// alır ve listeler, yazıcılar ise oluşturur, günceller ve siler.
//
// API sunucusuyla doğrudan konuşan yeni bir istemci oluşturmak için New fonksiyonu kullanılabilir.
//
// Kubernetes'te yaygın bir desen, bir önbellekten okumak ve API sunucusuna yazmaktır.
// Bu desen, bir Önbellek ile İstemci oluşturularak kapsanır.
//
// # Seçenekler
//
// Kubernetes'teki birçok istemci işlemi seçenekleri destekler. Bu seçenekler,
// belirli bir yöntem çağrısının sonunda değişken argümanlar olarak temsil edilir.
// Örneğin, bir liste üzerinde etiket seçici kullanmak için şu şekilde çağırabilirsiniz:
//
//	err := someReader.List(context.Background(), &podList, client.MatchingLabels{"somelabel": "someval"})
//
// # İndeksleme
//
// Önbelleklere bir FieldIndexer kullanarak indeksler eklenebilir. Bu, belirli özelliklere sahip
// nesneleri kolayca ve verimli bir şekilde aramanıza olanak tanır. Daha sonra, ilgili Önbelleğe
// karşılık gelen Okuyucu üzerindeki List çağrılarına bir alan seçici belirterek indeksi kullanabilirsiniz.
//
// and efficiently look up objects with certain properties.  You can then make
// use of the index by specifying a field selector on calls to List on the Reader
// corresponding to the given Cache.
//
// For instance, a Secret controller might have an index on the
// `.spec.volumes.secret.secretName` field in Pod objects, so that it could
// easily look up all pods that reference a given secret.
package client
