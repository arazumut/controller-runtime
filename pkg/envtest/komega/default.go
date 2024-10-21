package komega

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// defaultK, paket genelinde kullanılan Komega örneğidir.
var defaultK = &komega{ctx: context.Background()}

// SetClient, paket genelinde kullanılan client'ı ayarlar.
func SetClient(c client.Client) {
	defaultK.client = c
}

// SetContext, paket genelinde kullanılan context'i ayarlar.
func SetContext(c context.Context) {
	defaultK.ctx = c
}

// checkDefaultClient, defaultK.client'in ayarlanıp ayarlanmadığını kontrol eder.
func checkDefaultClient() {
	if defaultK.client == nil {
		panic("Default Komega'nın client'ı ayarlanmadı. SetClient ile ayarlayın.")
	}
}

// Get, bir kaynağı getirip oluşan hatayı döndüren bir fonksiyon döner.
// gomega.Eventually() ile şu şekilde kullanılabilir:
//
//	deployment := appsv1.Deployment{ ... }
//	gomega.Eventually(komega.Get(&deployment)).Should(gomega.Succeed())
//
// Dönen fonksiyon doğrudan çağrılarak şu şekilde de kullanılabilir: gomega.Expect(komega.Get(...)()).To(...)
func Get(obj client.Object) func() error {
	checkDefaultClient()
	return defaultK.Get(obj)
}

// List, kaynakları listeleyip oluşan hatayı döndüren bir fonksiyon döner.
// gomega.Eventually() ile şu şekilde kullanılabilir:
//
//	deployments := v1.DeploymentList{ ... }
//	gomega.Eventually(komega.List(&deployments)).Should(gomega.Succeed())
//
// Dönen fonksiyon doğrudan çağrılarak şu şekilde de kullanılabilir: gomega.Expect(komega.List(...)()).To(...)
func List(list client.ObjectList, opts ...client.ListOption) func() error {
	checkDefaultClient()
	return defaultK.List(list, opts...)
}

// Update, bir kaynağı getirip verilen güncelleme fonksiyonunu uygulayıp kaynağı güncelleyen bir fonksiyon döner.
// gomega.Eventually() ile şu şekilde kullanılabilir:
//
//	deployment := appsv1.Deployment{ ... }
//	gomega.Eventually(komega.Update(&deployment, func() {
//	  deployment.Spec.Replicas = 3
//	})).Should(gomega.Succeed())
//
// Dönen fonksiyon doğrudan çağrılarak şu şekilde de kullanılabilir: gomega.Expect(komega.Update(...)()).To(...)
func Update(obj client.Object, f func(), opts ...client.UpdateOption) func() error {
	checkDefaultClient()
	return defaultK.Update(obj, f, opts...)
}

// UpdateStatus, bir kaynağı getirip verilen güncelleme fonksiyonunu uygulayıp kaynağın durumunu güncelleyen bir fonksiyon döner.
// gomega.Eventually() ile şu şekilde kullanılabilir:
//
//	deployment := appsv1.Deployment{ ... }
//	gomega.Eventually(komega.UpdateStatus(&deployment, func() {
//	  deployment.Status.AvailableReplicas = 1
//	})).Should(gomega.Succeed())
//
// Dönen fonksiyon doğrudan çağrılarak şu şekilde de kullanılabilir: gomega.Expect(komega.UpdateStatus(...)()).To(...)
func UpdateStatus(obj client.Object, f func(), opts ...client.SubResourceUpdateOption) func() error {
	checkDefaultClient()
	return defaultK.UpdateStatus(obj, f, opts...)
}

// Object, bir kaynağı getirip objeyi döndüren bir fonksiyon döner.
// gomega.Eventually() ile şu şekilde kullanılabilir:
//
//	deployment := appsv1.Deployment{ ... }
//	gomega.Eventually(komega.Object(&deployment)).Should(HaveField("Spec.Replicas", gomega.Equal(ptr.To(3))))
//
// Dönen fonksiyon doğrudan çağrılarak şu şekilde de kullanılabilir: gomega.Expect(komega.Object(...)()).To(...)
func Object(obj client.Object) func() (client.Object, error) {
	checkDefaultClient()
	return defaultK.Object(obj)
}

// ObjectList, bir kaynağı getirip objeyi döndüren bir fonksiyon döner.
// gomega.Eventually() ile şu şekilde kullanılabilir:
//
//	deployments := appsv1.DeploymentList{ ... }
//	gomega.Eventually(komega.ObjectList(&deployments)).Should(HaveField("Items", HaveLen(1)))
//
// Dönen fonksiyon doğrudan çağrılarak şu şekilde de kullanılabilir: gomega.Expect(komega.ObjectList(...)()).To(...)
func ObjectList(list client.ObjectList, opts ...client.ListOption) func() (client.ObjectList, error) {
	checkDefaultClient()
	return defaultK.ObjectList(list, opts...)
}
