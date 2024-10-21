/*
2021 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa uyarınca veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMADAN.
Lisans kapsamındaki izinler ve sınırlamalar için Lisansa bakın.
*/

package cache

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"

	"sigs.k8s.io/controller-runtime/pkg/cache/internal"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllertest"
	crscheme "sigs.k8s.io/controller-runtime/pkg/scheme"
)

const (
	itemPointerSliceTypeGroupName = "jakob.fabian"
	itemPointerSliceTypeVersion   = "v1"
)

var _ = Describe("ip.objectTypeForListObject", func() {
	ip := &informerCache{
		scheme:    scheme.Scheme,
		Informers: &internal.Informers{},
	}

	It("yapılandırılmamış listeler için nesne türünü bulmalı", func() {
		unstructuredList := &unstructured.UnstructuredList{}
		unstructuredList.SetAPIVersion("v1")
		unstructuredList.SetKind("PodList")

		gvk, obj, err := ip.objectTypeForListObject(unstructuredList)
		Expect(err).ToNot(HaveOccurred())
		Expect(gvk.Group).To(Equal(""))
		Expect(gvk.Version).To(Equal("v1"))
		Expect(gvk.Kind).To(Equal("Pod"))
		referenceUnstructured := &unstructured.Unstructured{}
		referenceUnstructured.SetGroupVersionKind(*gvk)
		Expect(obj).To(Equal(referenceUnstructured))
	})

	It("kısmi nesne meta verileri listeleri için nesne türünü bulmalı", func() {
		partialList := &metav1.PartialObjectMetadataList{}
		partialList.APIVersion = ("v1")
		partialList.Kind = "PodList"

		gvk, obj, err := ip.objectTypeForListObject(partialList)
		Expect(err).ToNot(HaveOccurred())
		Expect(gvk.Group).To(Equal(""))
		Expect(gvk.Version).To(Equal("v1"))
		Expect(gvk.Kind).To(Equal("Pod"))
		referencePartial := &metav1.PartialObjectMetadata{}
		referencePartial.SetGroupVersionKind(*gvk)
		Expect(obj).To(Equal(referencePartial))
	})

	It("literal öğeler dilimi içeren bir listenin nesne türünü bulmalı", func() {
		gvk, obj, err := ip.objectTypeForListObject(&corev1.PodList{})
		Expect(err).ToNot(HaveOccurred())
		Expect(gvk.Group).To(Equal(""))
		Expect(gvk.Version).To(Equal("v1"))
		Expect(gvk.Kind).To(Equal("Pod"))
		referencePod := &corev1.Pod{}
		Expect(obj).To(Equal(referencePod))
	})

	It("işaretçi öğeler dilimi içeren bir listenin nesne türünü bulmalı", func() {
		By("türü kaydederek", func() {
			ip.scheme = runtime.NewScheme()
			err := (&crscheme.Builder{
				GroupVersion: schema.GroupVersion{Group: itemPointerSliceTypeGroupName, Version: itemPointerSliceTypeVersion},
			}).
				Register(
					&controllertest.UnconventionalListType{},
					&controllertest.UnconventionalListTypeList{},
				).AddToScheme(ip.scheme)
			Expect(err).ToNot(HaveOccurred())
		})

		By("objectTypeForListObject çağırarak", func() {
			gvk, obj, err := ip.objectTypeForListObject(&controllertest.UnconventionalListTypeList{})
			Expect(err).ToNot(HaveOccurred())
			Expect(gvk.Group).To(Equal(itemPointerSliceTypeGroupName))
			Expect(gvk.Version).To(Equal(itemPointerSliceTypeVersion))
			Expect(gvk.Kind).To(Equal("UnconventionalListType"))
			referenceObject := &controllertest.UnconventionalListType{}
			Expect(obj).To(Equal(referenceObject))
		})
	})
})
