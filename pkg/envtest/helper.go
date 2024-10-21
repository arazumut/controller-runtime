/*
2021 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
yetkiler ve sınırlamalar için Lisansa bakınız.
*/

package envtest

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

var (
	crdScheme = scheme.Scheme
)

// init, crdScheme paket değişkenini doğru şekilde başlatmak için gereklidir.
func init() {
	_ = apiextensionsv1.AddToScheme(crdScheme)
}

// mergePaths, iki dize dilimini birleştirir.
// Bu işlev, birleştirilmiş dilimin sırası hakkında garanti vermez.
func mergePaths(s1, s2 []string) []string {
	m := make(map[string]struct{})
	for _, s := range s1 {
		m[s] = struct{}{}
	}
	for _, s := range s2 {
		m[s] = struct{}{}
	}
	merged := make([]string, 0, len(m))
	for key := range m {
		merged = append(merged, key)
	}
	return merged
}

// mergeCRDs, iki CRD dilimini adlarını kullanarak birleştirir.
// Bu işlev, birleştirilmiş dilimin sırası hakkında garanti vermez.
func mergeCRDs(s1, s2 []*apiextensionsv1.CustomResourceDefinition) []*apiextensionsv1.CustomResourceDefinition {
	m := make(map[string]*apiextensionsv1.CustomResourceDefinition)
	for _, obj := range s1 {
		m[obj.GetName()] = obj
	}
	for _, obj := range s2 {
		m[obj.GetName()] = obj
	}
	merged := make([]*apiextensionsv1.CustomResourceDefinition, 0, len(m))
	for _, obj := range m {
		merged = append(merged, obj.DeepCopy())
	}
	return merged
}
