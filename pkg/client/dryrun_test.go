/*
2020 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa uyarınca veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ OLMAKSIZIN, açık veya zımni olarak.
Lisans kapsamında izin verilen belirli dil kapsamındaki
haklar ve sınırlamalar için Lisansa bakınız.
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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("DryRunClient", func() {
	var dep *appsv1.Deployment
	var count uint64 = 0
	var replicaCount int32 = 2
	var ns = "default"
	ctx := context.Background()

	getClient := func() client.Client {
		cl, err := client.New(cfg, client.Options{DryRun: ptr.To(true)})
		Expect(err).NotTo(HaveOccurred())
		Expect(cl).NotTo(BeNil())
		return cl
	}

	BeforeEach(func() {
		atomic.AddUint64(&count, 1)
		dep = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("dry-run-deployment-%v", count),
				Namespace: ns,
				Labels:    map[string]string{"name": fmt.Sprintf("dry-run-deployment-%v", count)},
			},
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

	It("bir nesneyi başarıyla Get yapmalı", func() {
		name := types.NamespacedName{Namespace: ns, Name: dep.Name}
		result := &appsv1.Deployment{}

		Expect(getClient().Get(ctx, name, result)).NotTo(HaveOccurred())
		Expect(result).To(BeEquivalentTo(dep))
	})

	It("nesneleri başarıyla Listelemeli", func() {
		result := &appsv1.DeploymentList{}
		opts := client.MatchingLabels(dep.Labels)

		Expect(getClient().List(ctx, result, opts)).NotTo(HaveOccurred())

		Expect(len(result.Items)).To(BeEquivalentTo(1))
		Expect(result.Items[0]).To(BeEquivalentTo(*dep))
	})

	It("bir nesne oluşturmamalı", func() {
		newDep := dep.DeepCopy()
		newDep.Name = "new-deployment"

		Expect(getClient().Create(ctx, newDep)).ToNot(HaveOccurred())

		_, err := clientset.AppsV1().Deployments(ns).Get(ctx, newDep.Name, metav1.GetOptions{})
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})

	It("opts ile bir nesne oluşturmamalı", func() {
		newDep := dep.DeepCopy()
		newDep.Name = "new-deployment"
		opts := &client.CreateOptions{DryRun: []string{"Bye", "Pippa"}}

		Expect(getClient().Create(ctx, newDep, opts)).ToNot(HaveOccurred())

		_, err := clientset.AppsV1().Deployments(ns).Get(ctx, newDep.Name, metav1.GetOptions{})
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})

	It("geçersiz bir nesne için oluşturma isteğini reddetmeli", func() {
		changedDep := dep.DeepCopy()
		changedDep.Spec.Template.Spec.Containers = nil

		err := getClient().Create(ctx, changedDep)
		Expect(apierrors.IsInvalid(err)).To(BeTrue())
	})

	It("güncelleme yoluyla nesneleri değiştirmemeli", func() {
		changedDep := dep.DeepCopy()
		*changedDep.Spec.Replicas = 2

		Expect(getClient().Update(ctx, changedDep)).ToNot(HaveOccurred())

		actual, err := clientset.AppsV1().Deployments(ns).Get(ctx, dep.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).NotTo(BeNil())
		Expect(actual).To(BeEquivalentTo(dep))
	})

	It("opts ile güncelleme yoluyla nesneleri değiştirmemeli", func() {
		changedDep := dep.DeepCopy()
		*changedDep.Spec.Replicas = 2
		opts := &client.UpdateOptions{DryRun: []string{"Bye", "Pippa"}}

		Expect(getClient().Update(ctx, changedDep, opts)).ToNot(HaveOccurred())

		actual, err := clientset.AppsV1().Deployments(ns).Get(ctx, dep.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).NotTo(BeNil())
		Expect(actual).To(BeEquivalentTo(dep))
	})

	It("geçersiz bir değişiklik için güncelleme isteğini reddetmeli", func() {
		changedDep := dep.DeepCopy()
		changedDep.Spec.Template.Spec.Containers = nil

		err := getClient().Update(ctx, changedDep)
		Expect(apierrors.IsInvalid(err)).To(BeTrue())
	})

	It("yama yoluyla nesneleri değiştirmemeli", func() {
		changedDep := dep.DeepCopy()
		*changedDep.Spec.Replicas = 2

		Expect(getClient().Patch(ctx, changedDep, client.MergeFrom(dep))).ToNot(HaveOccurred())

		actual, err := clientset.AppsV1().Deployments(ns).Get(ctx, dep.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).NotTo(BeNil())
		Expect(actual).To(BeEquivalentTo(dep))
	})

	It("opts ile yama yoluyla nesneleri değiştirmemeli", func() {
		changedDep := dep.DeepCopy()
		*changedDep.Spec.Replicas = 2
		opts := &client.PatchOptions{DryRun: []string{"Bye", "Pippa"}}

		Expect(getClient().Patch(ctx, changedDep, client.MergeFrom(dep), opts)).ToNot(HaveOccurred())

		actual, err := clientset.AppsV1().Deployments(ns).Get(ctx, dep.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).NotTo(BeNil())
		Expect(actual).To(BeEquivalentTo(dep))
	})

	It("nesneleri silmemeli", func() {
		Expect(getClient().Delete(ctx, dep)).NotTo(HaveOccurred())

		actual, err := clientset.AppsV1().Deployments(ns).Get(ctx, dep.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).NotTo(BeNil())
		Expect(actual).To(BeEquivalentTo(dep))
	})

	It("opts ile nesneleri silmemeli", func() {
		opts := &client.DeleteOptions{DryRun: []string{"Bye", "Pippa"}}

		Expect(getClient().Delete(ctx, dep, opts)).NotTo(HaveOccurred())

		actual, err := clientset.AppsV1().Deployments(ns).Get(ctx, dep.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).NotTo(BeNil())
		Expect(actual).To(BeEquivalentTo(dep))
	})

	It("deleteAllOf yoluyla nesneleri silmemeli", func() {
		opts := []client.DeleteAllOfOption{client.InNamespace(ns), client.MatchingLabels(dep.Labels)}

		Expect(getClient().DeleteAllOf(ctx, dep, opts...)).NotTo(HaveOccurred())

		actual, err := clientset.AppsV1().Deployments(ns).Get(ctx, dep.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).NotTo(BeNil())
		Expect(actual).To(BeEquivalentTo(dep))
	})

	It("durum güncellemesi yoluyla nesneleri değiştirmemeli", func() {
		changedDep := dep.DeepCopy()
		changedDep.Status.Replicas = 99

		Expect(getClient().Status().Update(ctx, changedDep)).NotTo(HaveOccurred())

		actual, err := clientset.AppsV1().Deployments(ns).Get(ctx, dep.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).NotTo(BeNil())
		Expect(actual).To(BeEquivalentTo(dep))
	})

	It("opts ile durum güncellemesi yoluyla nesneleri değiştirmemeli", func() {
		changedDep := dep.DeepCopy()
		changedDep.Status.Replicas = 99
		opts := &client.SubResourceUpdateOptions{UpdateOptions: client.UpdateOptions{DryRun: []string{"Bye", "Pippa"}}}

		Expect(getClient().Status().Update(ctx, changedDep, opts)).NotTo(HaveOccurred())

		actual, err := clientset.AppsV1().Deployments(ns).Get(ctx, dep.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).NotTo(BeNil())
		Expect(actual).To(BeEquivalentTo(dep))
	})

	It("durum yaması yoluyla nesneleri değiştirmemeli", func() {
		changedDep := dep.DeepCopy()
		changedDep.Status.Replicas = 99

		Expect(getClient().Status().Patch(ctx, changedDep, client.MergeFrom(dep))).ToNot(HaveOccurred())

		actual, err := clientset.AppsV1().Deployments(ns).Get(ctx, dep.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).NotTo(BeNil())
		Expect(actual).To(BeEquivalentTo(dep))
	})

	It("opts ile durum yaması yoluyla nesneleri değiştirmemeli", func() {
		changedDep := dep.DeepCopy()
		changedDep.Status.Replicas = 99

		opts := &client.SubResourcePatchOptions{PatchOptions: client.PatchOptions{DryRun: []string{"Bye", "Pippa"}}}

		Expect(getClient().Status().Patch(ctx, changedDep, client.MergeFrom(dep), opts)).ToNot(HaveOccurred())

		actual, err := clientset.AppsV1().Deployments(ns).Get(ctx, dep.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).NotTo(BeNil())
		Expect(actual).To(BeEquivalentTo(dep))
	})
})
