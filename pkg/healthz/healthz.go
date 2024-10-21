/*
2014 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
haklar ve sınırlamalar için Lisansa bakınız.
*/

package healthz

import (
	"fmt"
	"net/http"
	"path"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

// Handler, verilen denetleyicilerin sonuçlarını kök yola toplar ve
// denetleyicilerin adlarının alt yollarında bireysel denetleyicileri çağırmayı destekler.
//
// Denetimleri dinamik olarak eklemek *thread-safe* değildir -- bir sarmalayıcı kullanın.
type Handler struct {
	Checks map[string]Checker
}

// checkStatus belirli bir denetimin çıktısını tutar.
type checkStatus struct {
	name     string
	healthy  bool
	excluded bool
}

func (h *Handler) serveAggregated(resp http.ResponseWriter, req *http.Request) {
	failed := false
	excluded := getExcludedChecks(req)

	parts := make([]checkStatus, 0, len(h.Checks))

	// sonuçları hesapla...
	for checkName, check := range h.Checks {
		// denetimi hariç tutmak istiyorsak denetimi no-op yap
		if excluded.Has(checkName) {
			excluded.Delete(checkName)
			parts = append(parts, checkStatus{name: checkName, healthy: true, excluded: true})
			continue
		}
		if err := check(req); err != nil {
			log.V(1).Info("healthz denetimi başarısız", "denetleyici", checkName, "hata", err)
			parts = append(parts, checkStatus{name: checkName, healthy: false})
			failed = true
		} else {
			parts = append(parts, checkStatus{name: checkName, healthy: true})
		}
	}

	// ...hiç denetim yoksa varsayılan bir denetim ekle...
	if len(h.Checks) == 0 {
		parts = append(parts, checkStatus{name: "ping", healthy: true})
	}

	for _, c := range excluded.UnsortedList() {
		log.V(1).Info("sağlık denetimi hariç tutulamaz, eşleşme yok", "denetleyici", c)
	}

	// ...tutarlı olmak için sırala...
	sort.Slice(parts, func(i, j int) bool { return parts[i].name < parts[j].name })

	// ...ve sonucu yaz
	// TODO(directxman12): bu ayrıca JSON içeriği için bir istek kabul etmelidir (accept başlığı aracılığıyla)
	_, forceVerbose := req.URL.Query()["verbose"]
	writeStatusesAsText(resp, parts, excluded, failed, forceVerbose)
}

// writeStatusAsText verilen denetim durumlarını bazı yarı-keyfi
// özel metin formatında yazar. unknownExcludes, kullanıcının hariç tutulmasını istediği
// ancak aslında bilinen denetimler olmayan denetimleri listeler.
// writeStatusAsText başarısızlık durumunda her zaman ayrıntılıdır ve verilen argüman kullanılarak
// başarı durumunda ayrıntılı olmaya zorlanabilir.
func writeStatusesAsText(resp http.ResponseWriter, parts []checkStatus, unknownExcludes sets.Set[string], failed, forceVerbose bool) {
	resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	resp.Header().Set("X-Content-Type-Options", "nosniff")

	// her zaman önce durum kodunu yaz
	if failed {
		resp.WriteHeader(http.StatusInternalServerError)
	} else {
		resp.WriteHeader(http.StatusOK)
	}

	// kolay olmayan ayrıntılı başarı için kısayol
	if !failed && !forceVerbose {
		fmt.Fprint(resp, "ok")
		return
	}

	// başarısızlık durumunda her zaman ayrıntılıyız, bu yüzden bu noktadan itibaren ayrıntılı olmamız garanti

	for _, checkOut := range parts {
		switch {
		case checkOut.excluded:
			fmt.Fprintf(resp, "[+]%s hariç tutuldu: tamam\n", checkOut.name)
		case checkOut.healthy:
			fmt.Fprintf(resp, "[+]%s tamam\n", checkOut.name)
		default:
			// hatayı dahil etmeyin çünkü bu uç nokta herkese açıktır. Daha fazla ayrıntı isteyen biri
			// ayrıntılı denetimlere açıkça izin verilmelidir.
			fmt.Fprintf(resp, "[-]%s başarısız: sebep gizli\n", checkOut.name)
		}
	}

	if unknownExcludes.Len() > 0 {
		fmt.Fprintf(resp, "uyarı: bazı sağlık denetimleri hariç tutulamaz: %s için eşleşme yok\n", formatQuoted(unknownExcludes.UnsortedList()...))
	}

	if failed {
		log.Info("healthz denetimi başarısız", "durumlar", parts)
		fmt.Fprintf(resp, "healthz denetimi başarısız\n")
	} else {
		fmt.Fprint(resp, "healthz denetimi geçti\n")
	}
}

func (h *Handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	// isteği temizle (http.ServeMux'un iç mantığını biraz çoğaltarak)
	// yolu biraz temizle
	reqPath := req.URL.Path
	if reqPath == "" || reqPath[0] != '/' {
		reqPath = "/" + reqPath
	}
	// path.Clean kök hariç son eğik çizgiyi kaldırır
	// (bu sorun değil, çünkü yalnızca bir katman alt yolları sunuyoruz)
	reqPath = path.Clean(reqPath)

	// ya kök uç noktasına hizmet et...
	if reqPath == "/" {
		h.serveAggregated(resp, req)
		return
	}

	// ...hiçbir şey yoksa varsayılan denetim...
	if len(h.Checks) == 0 && reqPath[1:] == "ping" {
		CheckHandler{Checker: Ping}.ServeHTTP(resp, req)
		return
	}

	// ...veya bireysel bir denetleyici
	checkName := reqPath[1:] // öndeki eğik çizgiyi görmezden gel
	checker, known := h.Checks[checkName]
	if !known {
		http.NotFoundHandler().ServeHTTP(resp, req)
		return
	}

	CheckHandler{Checker: checker}.ServeHTTP(resp, req)
}

// CheckHandler, kök yolda bir sağlık denetimi uç noktası sunan bir http.Handler'dır,
// denetleyicisine dayalı olarak.
type CheckHandler struct {
	Checker
}

func (h CheckHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if err := h.Checker(req); err != nil {
		http.Error(resp, fmt.Sprintf("iç sunucu hatası: %v", err), http.StatusInternalServerError)
	} else {
		fmt.Fprint(resp, "tamam")
	}
}

// Checker, bir sağlık denetimi yapmayı bilir.
type Checker func(req *http.Request) error

// Ping kontrol edildiğinde otomatik olarak true döner.
var Ping Checker = func(_ *http.Request) error { return nil }

// getExcludedChecks, sorgu parametresinden hariç tutulacak sağlık denetim adlarını çıkarır.
func getExcludedChecks(r *http.Request) sets.Set[string] {
	checks, found := r.URL.Query()["exclude"]
	if found {
		return sets.New[string](checks...)
	}
	return sets.New[string]()
}

// formatQuoted, sağlık denetim adlarının biçimlendirilmiş bir dizesini döndürür,
// geçirilen sırayı koruyarak.
func formatQuoted(names ...string) string {
	quoted := make([]string, 0, len(names))
	for _, name := range names {
		quoted = append(quoted, fmt.Sprintf("%q", name))
	}
	return strings.Join(quoted, ",")
}
