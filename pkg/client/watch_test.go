/*
2020 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisansa bakınız.
*/

package client_test

import (
	"context"
	"fmt"
	"sync/atomic"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ClientWithWatch", func() {
	var dep *appsv1.Deployment
	var count uint64 = 0
	var replicaCount int32 = 2
	var ns = "kube-public"
	ctx := context.TODO()

	BeforeEach(func() {
		atomic.AddUint64(&count, 1)
		dep = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("watch-deployment-name-%v", count), Namespace: ns, Labels: map[string]string{"app": fmt.Sprintf("bar-%v", count)}},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicaCount,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"foo": "bar"},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"foo": "bar"}},
					Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "nginx", Image: "nginx"}}},
				},
			},
		}

		var err error
		dep, err = clientset.AppsV1().Deployments(ns).Create(ctx, dep, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		deleteDeployment(ctx, dep, ns)
	})

	Describe("NewWithWatch", func() {
		It("yeni bir Client döndürmeli", func() {
			cl, err := client.NewWithWatch(cfg, client.Options{})
			Expect(err).NotTo(HaveOccurred())
			Expect(cl).NotTo(BeNil())
		})

		watchSuite := func(through client.ObjectList, expectedType client.Object, checkGvk bool) {
			cl, err := client.NewWithWatch(cfg, client.Options{})
			Expect(err).NotTo(HaveOccurred())
			Expect(cl).NotTo(BeNil())

			watchInterface, err := cl.Watch(ctx, through, &client.ListOptions{
				FieldSelector: fields.OneTermEqualSelector("metadata.name", dep.Name),
				Namespace:     dep.Namespace,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(watchInterface).NotTo(BeNil())

			defer watchInterface.Stop()

			event, ok := <-watchInterface.ResultChan()
			Expect(ok).To(BeTrue())
			Expect(event.Type).To(BeIdenticalTo(watch.Added))
			Expect(event.Object).To(BeAssignableToTypeOf(expectedType))

			metaObject, ok := event.Object.(metav1.Object)
			Expect(ok).To(BeTrue())
			Expect(metaObject.GetName()).To(Equal(dep.Name))
			Expect(metaObject.GetUID()).To(Equal(dep.UID))

			if checkGvk {
				runtimeObject := event.Object
				gvk := runtimeObject.GetObjectKind().GroupVersionKind()
				Expect(gvk).To(Equal(schema.GroupVersionKind{
					Group:   "apps",
					Kind:    "Deployment",
					Version: "v1",
				}))
			}
		}

		It("tipli nesneyi izlerken bir oluşturma olayı almalı", func() {
			watchSuite(&appsv1.DeploymentList{}, &appsv1.Deployment{}, false)
		})

		It("yapılandırılmamış nesneyi izlerken bir oluşturma olayı almalı", func() {
			u := &unstructured.UnstructuredList{}
			u.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "apps",
				Kind:    "Deployment",
				Version: "v1",
			})
			watchSuite(u, &unstructured.Unstructured{}, true)
		})

		It("meta veri nesnesini izlerken bir oluşturma olayı almalı", func() {
			m := &metav1.PartialObjectMetadataList{TypeMeta: metav1.TypeMeta{Kind: "Deployment", APIVersion: "apps/v1"}}
			watchSuite(m, &metav1.PartialObjectMetadata{}, false)
		})
	})

})
