/*
2021 Kubernetes Yazarları tarafından telif hakkı saklıdır.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

    http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni olarak.
Lisans kapsamında izin verilen belirli dil altındaki haklar ve
sınırlamalar için Lisansa bakınız.
*/

// Bu paket, k8s.io/kubernetes/pkg/util/flock adresinden kopyalanmıştır
// ve k8s.io/kubernetes'i bir bağımlılık olarak eklememek için kullanılır.
//
// Unix sistemlerinde dosya kilitleme işlevsellikleri sağlar.
package flock
