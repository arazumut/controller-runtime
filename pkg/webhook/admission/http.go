/*
2018 Kubernetes Yazarları tarafından oluşturulmuştur.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMADAN, açık veya zımni.
Lisans kapsamındaki belirli dil izinleri ve sınırlamaları için
Lisans'a bakınız.
*/

package admission

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	v1 "k8s.io/api/admission/v1"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var admissionScheme = runtime.NewScheme()
var admissionCodecs = serializer.NewCodecFactory(admissionScheme)

// https://github.com/kubernetes/kubernetes/blob/c28c2009181fcc44c5f6b47e10e62dacf53e4da0/staging/src/k8s.io/pod-security-admission/cmd/webhook/server/server.go adresinden uyarlanmıştır.
//
// https://github.com/kubernetes/apiserver/blob/d6876a0600de06fef75968c4641c64d7da499f25/pkg/server/config.go#L433-L442C5 adresinden:
//
//	     1.5MB, etcd sunucusunun kabul etmesi gereken önerilen istemci istek boyutudur. Bkz.
//		 https://github.com/etcd-io/etcd/blob/release-3.4/embed/config.go#L56.
//		 Bir istek gövdesi json olarak kodlanmış olabilir ve etcd'de saklandığında proto'ya dönüştürülür,
//		 bu nedenle yazma isteğinde kabul edilip kodlanacak en büyük istek gövdesi boyutu olarak 2x izin veriyoruz.
//
// Kabul isteği için, kabul edilen nesnenin eski ve yeni sürümlerinin her birinin en fazla 3MB boyutunda olabileceğini
// ve isteğin geri kalanının 1MB'den az olacağını varsayabiliriz. Bu nedenle, maksimum istek boyutunu 7MB olarak ayarlayabiliriz.
// Kullanım durumunuz daha büyük maksimum istek boyutları gerektiriyorsa, lütfen bir sorun açın (https://github.com/kubernetes-sigs/controller-runtime/issues/new).
const maxRequestSize = int64(7 * 1024 * 1024)

func init() {
	utilruntime.Must(v1.AddToScheme(admissionScheme))
	utilruntime.Must(v1beta1.AddToScheme(admissionScheme))
}

var _ http.Handler = &Webhook{}

func (wh *Webhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if wh.WithContextFunc != nil {
		ctx = wh.WithContextFunc(ctx, r)
	}

	if r.Body == nil || r.Body == http.NoBody {
		err := errors.New("istek gövdesi boş")
		wh.getLogger(nil).Error(err, "kötü istek")
		wh.writeResponse(w, Errored(http.StatusBadRequest, err))
		return
	}

	defer r.Body.Close()
	limitedReader := &io.LimitedReader{R: r.Body, N: maxRequestSize}
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		wh.getLogger(nil).Error(err, "gelen isteğin gövdesini okuyamıyor")
		wh.writeResponse(w, Errored(http.StatusBadRequest, err))
		return
	}
	if limitedReader.N <= 0 {
		err := fmt.Errorf("istek varlığı çok büyük; limit %d bayt", maxRequestSize)
		wh.getLogger(nil).Error(err, "gelen isteğin gövdesini okuyamıyor; limit aşıldı")
		wh.writeResponse(w, Errored(http.StatusRequestEntityTooLarge, err))
		return
	}

	// içerik türünün doğru olduğunu doğrula
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		err = fmt.Errorf("contentType=%s, beklenen application/json", contentType)
		wh.getLogger(nil).Error(err, "bilinmeyen içerik türü ile bir isteği işleyemiyor")
		wh.writeResponse(w, Errored(http.StatusBadRequest, err))
		return
	}

	// Hem v1 hem de v1beta1 AdmissionReview türleri tamamen aynıdır, bu nedenle v1beta1 türü v1 türüne kodlanabilir.
	// Ancak runtime codec'in decoder'ı, bir Object'in TypeMeta'sı ayarlanmadıysa tür adından hangi türe kodlanacağını tahmin eder.
	// Kayıtlı olmayan bir türün TypeMeta'sını v1 GVK'ya ayarlayarak, decoder bir v1beta1 AdmissionReview'ı v1'e zorlar.
	// Gerçek AdmissionReview GVK, birden fazla sürüme izin verilmişse yazılı bir yanıt oluşturmak için kullanılacaktır,
	// aksi takdirde bu yanıt başarısız olacaktır.
	req := Request{}
	ar := unversionedAdmissionReview{}
	// ekstra bir kopyadan kaçının
	ar.Request = &req.AdmissionRequest
	ar.SetGroupVersionKind(v1.SchemeGroupVersion.WithKind("AdmissionReview"))
	_, actualAdmRevGVK, err := admissionCodecs.UniversalDeserializer().Decode(body, nil, &ar)
	if err != nil {
		wh.getLogger(nil).Error(err, "isteği kodlayamıyor")
		wh.writeResponse(w, Errored(http.StatusBadRequest, err))
		return
	}
	wh.getLogger(&req).V(5).Info("istek alındı")

	wh.writeResponseTyped(w, wh.Handle(ctx, req), actualAdmRevGVK)
}

// writeResponse, yanıtı w'ye genel olarak yazar, yani GVK bilgilerini kodlamadan.
func (wh *Webhook) writeResponse(w io.Writer, response Response) {
	wh.writeAdmissionResponse(w, v1.AdmissionReview{Response: &response.AdmissionResponse})
}

// writeResponseTyped, yanıtı w'ye admRevGVK'ya ayarlanmış GVK ile yazar, bu, webhook tarafından birden fazla AdmissionReview sürümüne izin verilmişse gereklidir.
func (wh *Webhook) writeResponseTyped(w io.Writer, response Response, admRevGVK *schema.GroupVersionKind) {
	ar := v1.AdmissionReview{
		Response: &response.AdmissionResponse,
	}
	// Varsayılan olarak bir v1 AdmissionReview kullanın, aksi takdirde API sunucusu istekleri tanımayabilir
	// webhook yapılandırması tarafından birden fazla AdmissionReview sürümüne izin verilmişse.
	// TODO: bu yapılandırılabilir olmalıdır çünkü eski API sunucuları v1'i bilmeyecektir.
	if admRevGVK == nil || *admRevGVK == (schema.GroupVersionKind{}) {
		ar.SetGroupVersionKind(v1.SchemeGroupVersion.WithKind("AdmissionReview"))
	} else {
		ar.SetGroupVersionKind(*admRevGVK)
	}
	wh.writeAdmissionResponse(w, ar)
}

// writeAdmissionResponse, ar'yi w'ye yazar.
func (wh *Webhook) writeAdmissionResponse(w io.Writer, ar v1.AdmissionReview) {
	if err := json.NewEncoder(w).Encode(ar); err != nil {
		wh.getLogger(nil).Error(err, "yanıtı kodlayamıyor ve yazamıyor")
		// `ar v1.AdmissionReview` açık ve yasal bir nesne olduğundan,
		// baytlara dönüştürülmesinde sorun olmamalıdır.
		// Buradaki hata muhtemelen anormal HTTP bağlantısından kaynaklanmaktadır,
		// örneğin, kırık boru, bu nedenle hata yanıtını yalnızca bir kez yazabiliriz,
		// sonsuz döngüsel çağırmayı önlemek için.
		serverError := Errored(http.StatusInternalServerError, err)
		if err = json.NewEncoder(w).Encode(v1.AdmissionReview{Response: &serverError.AdmissionResponse}); err != nil {
			wh.getLogger(nil).Error(err, "hala InternalServerError yanıtını kodlayamıyor ve yazamıyor")
		}
	} else {
		res := ar.Response
		if log := wh.getLogger(nil); log.V(5).Enabled() {
			if res.Result != nil {
				log = log.WithValues("kod", res.Result.Code, "sebep", res.Result.Reason, "mesaj", res.Result.Message)
			}
			log.V(5).Info("yanıt yazıldı", "istekID", res.UID, "izin verildi", res.Allowed)
		}
	}
}

// unversionedAdmissionReview, hem v1 hem de v1beta1 AdmissionReview türlerini kodlamak için kullanılır.
type unversionedAdmissionReview struct {
	v1.AdmissionReview
}

var _ runtime.Object = &unversionedAdmissionReview{}
