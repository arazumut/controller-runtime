/*
2019 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") kapsamında lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
haklar ve sınırlamalar için Lisans'a bakınız.
*/

package webhook

import (
	"gomodules.xyz/jsonpatch/v2"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// webhook işlevselliğinin yaygın parçaları için bazı takma adlar tanımlayın

// CustomDefaulter, kaynaklar üzerinde varsayılan ayarları belirlemek için işlevler tanımlar.
type CustomDefaulter = admission.CustomDefaulter

// CustomValidator, bir işlemi doğrulamak için işlevler tanımlar.
type CustomValidator = admission.CustomValidator

// AdmissionRequest, bir kabul işleyicisi için girişi tanımlar.
// Nesneyi tanımlamak için bilgi içerir (grup, sürüm, tür, kaynak, alt kaynak,
// ad, ad alanı) ve ayrıca ilgili işlemi (ör. Get, Create, vb.) ve nesnenin kendisini içerir.
type AdmissionRequest = admission.Request

// AdmissionResponse, bir kabul işleyicisinin çıktısıdır.
// Belirli bir işlemin izin verilip verilmediğini belirten bir yanıt içerir
// ve mutasyon kabul işleyicisi durumunda nesneyi değiştirmek için bir dizi yama içerir.
type AdmissionResponse = admission.Response

// Admission, sunucuya kayıt için uygun bir webhook'tur
// API işlemlerini doğrulayan ve potansiyel olarak içeriklerini değiştiren bir kabul webhook'udur.
type Admission = admission.Webhook

// AdmissionHandler, kabul isteklerini nasıl işleyeceğini bilen,
// onları doğrulayan ve potansiyel olarak içerdiği nesneleri değiştiren bir işleyicidir.
type AdmissionHandler = admission.Handler

// AdmissionDecoder, kabul isteklerinden nesneleri nasıl çözeceğini bilir.
type AdmissionDecoder = admission.Decoder

// JSONPatchOp, tek bir JSONPatch yama işlemini temsil eder.
type JSONPatchOp = jsonpatch.Operation

var (
	// Allowed, kabul isteğinin verilen neden için izin verilmesi gerektiğini belirtir.
	Allowed = admission.Allowed

	// Denied, kabul isteğinin verilen neden için reddedilmesi gerektiğini belirtir.
	Denied = admission.Denied

	// Patched, kabul isteğinin verilen neden için izin verilmesi gerektiğini
	// ve içerilen nesnenin verilen yamalar kullanılarak değiştirilmesi gerektiğini belirtir.
	Patched = admission.Patched

	// Errored, kabul isteğinde bir hata oluştuğunu belirtir.
	Errored = admission.Errored
)
