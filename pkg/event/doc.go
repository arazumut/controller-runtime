/*
2018 Kubernetes Yazarları tarafından oluşturulmuştur.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

    http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izinle zorunlu kılınmadıkça,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ OLMAKSIZIN, açık veya zımni olarak.
Lisans kapsamındaki izinler ve sınırlamalar için Lisansa bakınız.
*/

/*
Event paketinde, source.Sources tarafından üretilen ve handler.EventHandler tarafından
reconcile.Requests'e dönüştürülen Event türlerinin tanımları bulunur.

Bu türlerle doğrudan çalışmanız nadiren gerekecektir -- bunun yerine,
source.Sources ve handler.EventHandlers ile Controller.Watch kullanın.

Olaylar genellikle olaya neden olan tam bir runtime.Object ve
bu nesnenin meta verilerine doğrudan bir bağlantı içerir. Bu, Olaylarla çalışan
kodda çok fazla tür dönüştürmeyi önler.
*/
package event
