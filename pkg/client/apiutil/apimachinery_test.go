/*
2024 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN.
Lisans kapsamındaki izinler ve sınırlamalar hakkında daha fazla bilgi için Lisansı inceleyin.
*/

package apiutil_test

import (
	"context"
	"strconv"
	"testing"

	gmg "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func TestApiMachinery(t *testing.T) {
	for _, aggregatedDiscovery := range []bool{true, false} {
		t.Run("aggregatedDiscovery="+strconv.FormatBool(aggregatedDiscovery), func(t *testing.T) {
			restCfg := setupEnvtest(t, !aggregatedDiscovery)

			// Başlangıçta kaydedilen GVK'nın detayları.
			initialGvk := metav1.GroupVersionKind{
				Group:   "crew.example.com",
				Version: "v1",
				Kind:    "Driver",
			}

			// Çeşitli özelliklere sahip GVK'ları çalışma zamanında kaydetmek için bir dizi.
			runtimeGvks := []struct {
				name   string
				gvk    metav1.GroupVersionKind
				plural string
			}{
				{
					name: "Mevcut Gruba yeni Tür ve Sürüm eklendi",
					gvk: metav1.GroupVersionKind{
						Group:   "crew.example.com",
						Version: "v1alpha1",
						Kind:    "Passenger",
					},
					plural: "passengers",
				},
				{
					name: "Mevcut Grup ve Sürüme yeni Tür eklendi",
					gvk: metav1.GroupVersionKind{
						Group:   "crew.example.com",
						Version: "v1",
						Kind:    "Garage",
					},
					plural: "garages",
				},
				{
					name: "Yeni GVK",
					gvk: metav1.GroupVersionKind{
						Group:   "inventory.example.com",
						Version: "v1",
						Kind:    "Taxi",
					},
					plural: "taxis",
				},
			}

			t.Run("IsGVKNamespaced başlangıçta kaydedilen GVK için kapsamı rapor etmelidir", func(t *testing.T) {
				g := gmg.NewWithT(t)

				httpClient, err := rest.HTTPClientFor(restCfg)
				g.Expect(err).NotTo(gmg.HaveOccurred())

				lazyRestMapper, err := apiutil.NewDynamicRESTMapper(restCfg, httpClient)
				g.Expect(err).NotTo(gmg.HaveOccurred())

				s := scheme.Scheme
				err = apiextensionsv1.AddToScheme(s)
				g.Expect(err).NotTo(gmg.HaveOccurred())

				// Başlangıçta kaydedilen bir GVK'nın kapsamını sorgula.
				scope, err := apiutil.IsGVKNamespaced(
					schema.GroupVersionKind(initialGvk),
					lazyRestMapper,
				)
				g.Expect(err).NotTo(gmg.HaveOccurred())
				g.Expect(scope).To(gmg.BeTrue())
			})

			for _, runtimeGvk := range runtimeGvks {
				t.Run("IsGVKNamespaced "+runtimeGvk.name+" için kapsamı rapor etmelidir", func(t *testing.T) {
					g := gmg.NewWithT(t)
					ctx := context.Background()

					httpClient, err := rest.HTTPClientFor(restCfg)
					g.Expect(err).NotTo(gmg.HaveOccurred())

					lazyRestMapper, err := apiutil.NewDynamicRESTMapper(restCfg, httpClient)
					g.Expect(err).NotTo(gmg.HaveOccurred())

					s := scheme.Scheme
					err = apiextensionsv1.AddToScheme(s)
					g.Expect(err).NotTo(gmg.HaveOccurred())

					c, err := client.New(restCfg, client.Options{Scheme: s})
					g.Expect(err).NotTo(gmg.HaveOccurred())

					// Geçerli bir sorgu çalıştırarak önbelleği başlat.
					scope, err := apiutil.IsGVKNamespaced(
						schema.GroupVersionKind(initialGvk),
						lazyRestMapper,
					)
					g.Expect(err).NotTo(gmg.HaveOccurred())
					g.Expect(scope).To(gmg.BeTrue())

					// Çalışma zamanında yeni bir CRD kaydet.
					crd := newCRD(ctx, g, c, runtimeGvk.gvk.Group, runtimeGvk.gvk.Kind, runtimeGvk.plural)
					version := crd.Spec.Versions[0]
					version.Name = runtimeGvk.gvk.Version
					version.Storage = true
					version.Served = true
					crd.Spec.Versions = []apiextensionsv1.CustomResourceDefinitionVersion{version}
					crd.Spec.Scope = apiextensionsv1.NamespaceScoped

					g.Expect(c.Create(ctx, crd)).To(gmg.Succeed())
					t.Cleanup(func() {
						g.Expect(c.Delete(ctx, crd)).To(gmg.Succeed())
					})

					// CRD'nin kaydedilmesini bekle.
					g.Eventually(func(g gmg.Gomega) {
						isRegistered, err := isCrdRegistered(restCfg, runtimeGvk.gvk)
						g.Expect(err).NotTo(gmg.HaveOccurred())
						g.Expect(isRegistered).To(gmg.BeTrue())
					}).Should(gmg.Succeed(), "GVK mevcut olmalı")

					// Çalışma zamanında kaydedilen GVK'nın kapsamını sorgula.
					scope, err = apiutil.IsGVKNamespaced(
						schema.GroupVersionKind(runtimeGvk.gvk),
						lazyRestMapper,
					)
					g.Expect(err).NotTo(gmg.HaveOccurred())
					g.Expect(scope).To(gmg.BeTrue())
				})
			}
		})
	}
}

// Bir APIResource diliminde belirli bir Türün olup olmadığını kontrol et.
func kindInAPIResources(resources *metav1.APIResourceList, kind string) bool {
	for _, res := range resources.APIResources {
		if res.Kind == kind {
			return true
		}
	}
	return false
}

// Bir CRD'nin API sunucusuna DiscoveryClient kullanarak kaydedilip kaydedilmediğini kontrol et.
func isCrdRegistered(cfg *rest.Config, gvk metav1.GroupVersionKind) (bool, error) {
	discHTTP, err := rest.HTTPClientFor(cfg)
	if err != nil {
		return false, err
	}

	discClient, err := discovery.NewDiscoveryClientForConfigAndClient(cfg, discHTTP)
	if err != nil {
		return false, err
	}

	resources, err := discClient.ServerResourcesForGroupVersion(gvk.Group + "/" + gvk.Version)
	if err != nil {
		return false, err
	}

	return kindInAPIResources(resources, gvk.Kind), nil
}
