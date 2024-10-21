/*
2021 Kubernetes Yazarları tarafından oluşturulmuştur.

Apache Lisansı, Sürüm 2.0 ("Lisans") altında lisanslanmıştır;
bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
bu yazılım Lisans kapsamında "OLDUĞU GİBİ" dağıtılmakta olup,
HERHANGİ BİR GARANTİ VERİLMEMEKTEDİR; ne açık ne de zımni.
Lisans kapsamındaki izinler ve sınırlamalar hakkında daha fazla bilgi için Lisansı inceleyin.
*/

package komega

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Komega, sahte bir Kubernetes API'si içeren testler yazmak için bir dizi yardımcı programdır.
type Komega interface {
	// Get, bir kaynağı getiren ve oluşan hatayı döndüren bir fonksiyon döndürür.
	// Bu, gomega.Eventually() ile şu şekilde kullanılabilir:
	//   deployment := appsv1.Deployment{ ... }
	//   gomega.Eventually(k.Get(&deployment)).To(gomega.Succeed())
	// Döndürülen fonksiyon doğrudan çağrılarak gomega.Expect(k.Get(...)()).To(...) ile de kullanılabilir.
	Get(client.Object) func() error

	// List, kaynakları listeleyen ve oluşan hatayı döndüren bir fonksiyon döndürür.
	// Bu, gomega.Eventually() ile şu şekilde kullanılabilir:
	//   deployments := v1.DeploymentList{ ... }
	//   gomega.Eventually(k.List(&deployments)).To(gomega.Succeed())
	// Döndürülen fonksiyon doğrudan çağrılarak gomega.Expect(k.List(...)()).To(...) ile de kullanılabilir.
	List(client.ObjectList, ...client.ListOption) func() error

	// Update, bir kaynağı getiren, sağlanan güncelleme fonksiyonunu uygulayan ve ardından kaynağı güncelleyen bir fonksiyon döndürür.
	// Bu, gomega.Eventually() ile şu şekilde kullanılabilir:
	//   deployment := appsv1.Deployment{ ... }
	//   gomega.Eventually(k.Update(&deployment, func() {
	//     deployment.Spec.Replicas = 3
	//   })).To(gomega.Succeed())
	// Döndürülen fonksiyon doğrudan çağrılarak gomega.Expect(k.Update(...)()).To(...) ile de kullanılabilir.
	Update(client.Object, func(), ...client.UpdateOption) func() error

	// UpdateStatus, bir kaynağı getiren, sağlanan güncelleme fonksiyonunu uygulayan ve ardından kaynağın durumunu güncelleyen bir fonksiyon döndürür.
	// Bu, gomega.Eventually() ile şu şekilde kullanılabilir:
	//   deployment := appsv1.Deployment{ ... }
	//   gomega.Eventually(k.UpdateStatus(&deployment, func() {
	//     deployment.Status.AvailableReplicas = 1
	//   })).To(gomega.Succeed())
	// Döndürülen fonksiyon doğrudan çağrılarak gomega.Expect(k.UpdateStatus(...)()).To(...) ile de kullanılabilir.
	UpdateStatus(client.Object, func(), ...client.SubResourceUpdateOption) func() error

	// Object, bir kaynağı getiren ve nesneyi döndüren bir fonksiyon döndürür.
	// Bu, gomega.Eventually() ile şu şekilde kullanılabilir:
	//   deployment := appsv1.Deployment{ ... }
	//   gomega.Eventually(k.Object(&deployment)).To(HaveField("Spec.Replicas", gomega.Equal(ptr.To(int32(3)))))
	// Döndürülen fonksiyon doğrudan çağrılarak gomega.Expect(k.Object(...)()).To(...) ile de kullanılabilir.
	Object(client.Object) func() (client.Object, error)

	// ObjectList, bir kaynağı getiren ve nesneyi döndüren bir fonksiyon döndürür.
	// Bu, gomega.Eventually() ile şu şekilde kullanılabilir:
	//   deployments := appsv1.DeploymentList{ ... }
	//   gomega.Eventually(k.ObjectList(&deployments)).To(HaveField("Items", HaveLen(1)))
	// Döndürülen fonksiyon doğrudan çağrılarak gomega.Expect(k.ObjectList(...)()).To(...) ile de kullanılabilir.
	ObjectList(client.ObjectList, ...client.ListOption) func() (client.ObjectList, error)

	// WithContext, verilen bağlamı kullanan bir kopya döndürür.
	WithContext(context.Context) Komega
}
