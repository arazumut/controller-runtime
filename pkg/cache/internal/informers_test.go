/*
2022 Kubernetes Yazarları tarafından oluşturulmuştur.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
Lisans'ın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izinle gerekli olmadıkça,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN.
Lisans kapsamındaki izinler ve sınırlamalar hakkında daha fazla bilgi için
Lisans'a bakınız.
*/

package internal

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
)

// gvkFixupWatcher'ın watch.FakeWatcher gibi davrandığını
// ve GVK'yı (GroupVersionKind) geçersiz kıldığını test edin.
// Bu testler, watch.FakeWatcher testlerinden uyarlanmıştır:
// https://github.com/kubernetes/kubernetes/blob/adbda068c1808fcc8a64a94269e0766b5c46ec41/staging/src/k8s.io/apimachinery/pkg/watch/watch_test.go#L33-L78
var _ = Describe("gvkFixupWatcher", func() {
	It("watch.FakeWatcher gibi davranır", func() {
		// Yeni bir test tipi oluşturur
		newTestType := func(name string) runtime.Object {
			return &metav1.PartialObjectMetadata{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
			}
		}

		f := watch.NewFake()
		// Bu, sarmalayıcının tüm olaylara ayarlamasını beklediğimiz GVK'dır
		expectedGVK := schema.GroupVersionKind{
			Group:   "testgroup",
			Version: "v1test2",
			Kind:    "TestKind",
		}
		gvkfw := newGVKFixupWatcher(expectedGVK, f)

		// Test tablosu
		table := []struct {
			t watch.EventType
			s runtime.Object
		}{
			{watch.Added, newTestType("foo")},
			{watch.Modified, newTestType("qux")},
			{watch.Modified, newTestType("bar")},
			{watch.Deleted, newTestType("bar")},
			{watch.Error, newTestType("error: blah")},
		}

		// Tüketici fonksiyonu
		consumer := func(w watch.Interface) {
			for _, expect := range table {
				By(fmt.Sprintf("watch.EventType: %v'yi düzeltiyor ve iletiyor", expect.t))
				got, ok := <-w.ResultChan()
				Expect(ok).To(BeTrue(), "erken kapandı")
				Expect(expect.t).To(Equal(got.Type), "beklenmeyen Event.Type veya sırasız Event")
				Expect(got.Object).To(BeAssignableToTypeOf(&metav1.PartialObjectMetadata{}), "beklenmeyen Event.Object türü")
				a := got.Object.(*metav1.PartialObjectMetadata)
				Expect(got.Object.GetObjectKind().GroupVersionKind()).To(Equal(expectedGVK), "GVK düzeltilmedi")
				expected := expect.s.DeepCopyObject()
				expected.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{})
				actual := a.DeepCopyObject()
				actual.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{})
				Expect(actual).To(Equal(expected), "Nesnede beklenmeyen değişiklik")
			}
			Eventually(w.ResultChan()).Should(BeClosed())
		}

		// Gönderici fonksiyonu
		sender := func() {
			f.Add(newTestType("foo"))
			f.Action(watch.Modified, newTestType("qux"))
			f.Modify(newTestType("bar"))
			f.Delete(newTestType("bar"))
			f.Error(newTestType("error: blah"))
			f.Stop()
		}

		go sender()
		consumer(gvkfw)
	})
})
