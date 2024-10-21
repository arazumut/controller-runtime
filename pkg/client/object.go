/*
2020 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ OLMAKSIZIN; açık veya zımni garantiler dahil ancak bunlarla sınırlı olmamak üzere.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisansa bakın.
*/

package client

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Object, Kubernetes nesnesidir, herhangi bir kaynağın her iki Object arayüzünü de
// uygulayan işlevlerle ayrım gözetmeksizin çalışmasına olanak tanır.
//
// Anlamsal olarak, bunlar hem serileştirilebilir (runtime.Object) hem de tanımlanabilir
// (metav1.Object) nesnelerdir -- YAML veya JSON olarak yazabileceğiniz ve ardından
// `kubectl create` komutunu çalıştırabileceğiniz herhangi bir nesne düşünün.
//
// Kod açısından, ObjectMeta'yı (metav1.Object sağlar) ve TypeMeta'yı (runtime.Object'in
// yarısını sağlar) içeren ve bir `DeepCopyObject` uygulamasına sahip (runtime.Object'in
// diğer yarısı) herhangi bir nesne varsayılan olarak bunu uygular.
//
// Örneğin, neredeyse tüm yerleşik türler Objects'dir ve ayrıca tüm KubeBuilder tarafından
// oluşturulan CRD'ler (onlara gerçekten tuhaf bir şey yapmadığınız sürece).
//
// Büyük ölçüde, runtime.Object'i uygulayan çoğu şey aynı zamanda Object'i de uygular --
// sadece bir runtime.Object uygulamasına sahip olmak çok nadirdir (durumlar genellikle
// `metadata` alanına sahip olmayan Webhook yükleri gibi tuhaf yerleşik türlerdir).
//
// XYZList türlerinin farklı olduğunu unutmayın: ObjectList'i uygularlar.
type Object interface {
	metav1.Object
	runtime.Object
}

// ObjectList, Kubernetes nesne listesidir, herhangi bir kaynağın hem runtime.Object hem de
// metav1.ListInterface arayüzlerini uygulayan işlevlerle ayrım gözetmeksizin çalışmasına
// olanak tanır.
//
// Anlamsal olarak, bu herhangi bir nesne serileştirilebilir (ObjectMeta) ve bir
// Kubernetes liste sarmalayıcısıdır (öğeler, sayfalama alanları, vb. içerir) --
// `kubectl list --output yaml` çağrısının yanıtında kullanılan sarmalayıcıyı düşünün.
//
// Kod açısından, ListMeta'yı (metav1.ListInterface sağlar) ve TypeMeta'yı (runtime.Object'in
// yarısını sağlar) içeren ve bir `DeepCopyObject` uygulamasına sahip (runtime.Object'in
// diğer yarısı) herhangi bir nesne varsayılan olarak bunu uygular.
//
// Örneğin, neredeyse tüm yerleşik XYZList türleri ObjectLists'dir ve ayrıca tüm
// KubeBuilder tarafından oluşturulan CRD'lerin XYZList türleri (onlara gerçekten tuhaf
// bir şey yapmadığınız sürece).
//
// Büyük ölçüde, XYZList olan ve runtime.Object'i uygulayan çoğu şey aynı zamanda ObjectList'i
// de uygular -- sadece bir runtime.Object uygulamasına sahip olmak çok nadirdir (durumlar
// genellikle `metadata` alanına sahip olmayan Webhook yükleri gibi tuhaf yerleşik türlerdir).
//
// Bu, listedeki öğelerin kendileri tarafından neredeyse her zaman uygulanan Object'e benzer.
type ObjectList interface {
	metav1.ListInterface
	runtime.Object
}
