/*
2021 Kubernetes Yazarları tarafından yazılmıştır.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamındaki izinler ve sınırlamalar hakkında daha fazla bilgi için
Lisans'a bakınız.
*/

package conversion

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

// Decoder, CRD sürüm dönüştürme isteğinin içeriğini somut bir nesneye
// nasıl kod çözeceğini bilir.
// TODO(droot): bu iş için admission paketindeki decoder'ı yeniden kullanmayı düşün.
type Decoder struct {
	codecs serializer.CodecFactory
}

// NewDecoder, verilen runtime.Scheme ile bir Decoder oluşturur
func NewDecoder(scheme *runtime.Scheme) *Decoder {
	if scheme == nil {
		panic("scheme asla nil olmamalı")
	}
	return &Decoder{codecs: serializer.NewCodecFactory(scheme)}
}

// Decode, iç içe geçmiş nesneyi çözer.
func (d *Decoder) Decode(content []byte) (runtime.Object, *schema.GroupVersionKind, error) {
	deserializer := d.codecs.UniversalDeserializer()
	return deserializer.Decode(content, nil, nil)
}

// DecodeInto, iç içe geçmiş nesneyi verilen runtime.Object içine çözer.
func (d *Decoder) DecodeInto(content []byte, into runtime.Object) error {
	deserializer := d.codecs.UniversalDeserializer()
	return runtime.DecodeInto(deserializer, content, into)
}
