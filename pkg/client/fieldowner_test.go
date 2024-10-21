/*
2024 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa gereği veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ OLMAKSIZIN, açık veya zımni.
Lisans kapsamındaki izinleri ve
sınırlamaları yöneten özel dil için Lisansa bakınız.
*/

package client_test

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestAlanSahibiIle(t *testing.T) {
	cagrilar := 0
	sahteClient := testClient(t, "ozel-alan-yoneticisi", func() { cagrilar++ })
	sarilmisClient := client.WithFieldOwner(sahteClient, "ozel-alan-yoneticisi")

	ctx := context.Background()
	sahteNesne := &corev1.Namespace{}

	_ = sarilmisClient.Create(ctx, sahteNesne)
	_ = sarilmisClient.Update(ctx, sahteNesne)
	_ = sarilmisClient.Patch(ctx, sahteNesne, nil)
	_ = sarilmisClient.Status().Create(ctx, sahteNesne, sahteNesne)
	_ = sarilmisClient.Status().Update(ctx, sahteNesne)
	_ = sarilmisClient.Status().Patch(ctx, sahteNesne, nil)
	_ = sarilmisClient.SubResource("bazi-altkaynak").Create(ctx, sahteNesne, sahteNesne)
	_ = sarilmisClient.SubResource("bazi-altkaynak").Update(ctx, sahteNesne)
	_ = sarilmisClient.SubResource("bazi-altkaynak").Patch(ctx, sahteNesne, nil)

	if beklenenCagrilar := 9; cagrilar != beklenenCagrilar {
		t.Fatalf("beklenen çağrı sayısı yanlış: beklenen=%d; elde edilen=%d", beklenenCagrilar, cagrilar)
	}
}

func TestAlanSahibiIleGecersizKilindi(t *testing.T) {
	cagrilar := 0

	sahteClient := testClient(t, "yeni-alan-yoneticisi", func() { cagrilar++ })
	sarilmisClient := client.WithFieldOwner(sahteClient, "eski-alan-yoneticisi")

	ctx := context.Background()
	sahteNesne := &corev1.Namespace{}

	_ = sarilmisClient.Create(ctx, sahteNesne, client.FieldOwner("yeni-alan-yoneticisi"))
	_ = sarilmisClient.Update(ctx, sahteNesne, client.FieldOwner("yeni-alan-yoneticisi"))
	_ = sarilmisClient.Patch(ctx, sahteNesne, nil, client.FieldOwner("yeni-alan-yoneticisi"))
	_ = sarilmisClient.Status().Create(ctx, sahteNesne, sahteNesne, client.FieldOwner("yeni-alan-yoneticisi"))
	_ = sarilmisClient.Status().Update(ctx, sahteNesne, client.FieldOwner("yeni-alan-yoneticisi"))
	_ = sarilmisClient.Status().Patch(ctx, sahteNesne, nil, client.FieldOwner("yeni-alan-yoneticisi"))
	_ = sarilmisClient.SubResource("bazi-altkaynak").Create(ctx, sahteNesne, sahteNesne, client.FieldOwner("yeni-alan-yoneticisi"))
	_ = sarilmisClient.SubResource("bazi-altkaynak").Update(ctx, sahteNesne, client.FieldOwner("yeni-alan-yoneticisi"))
	_ = sarilmisClient.SubResource("bazi-altkaynak").Patch(ctx, sahteNesne, nil, client.FieldOwner("yeni-alan-yoneticisi"))

	if beklenenCagrilar := 9; cagrilar != beklenenCagrilar {
		t.Fatalf("beklenen çağrı sayısı yanlış: beklenen=%d; elde edilen=%d", beklenenCagrilar, cagrilar)
	}
}

// testClient, çağrıların beklenen alan yöneticisine sahip olup olmadığını kontrol eden
// ve her yakalanan çağrıda geri çağırma fonksiyonunu çağıran yardımcı bir fonksiyondur.
func testClient(t *testing.T, beklenenAlanYoneticisi string, geriCagir func()) client.Client {
	// TODO: interceptor paketindeki dummyClient'i kullanabiliriz eğer onu bir internal pakete taşırız
	return fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
		Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
			geriCagir()
			out := &client.CreateOptions{}
			for _, f := range opts {
				f.ApplyToCreate(out)
			}
			if eldeEdilen := out.FieldManager; beklenenAlanYoneticisi != eldeEdilen {
				t.Fatalf("yanlış alan yöneticisi: beklenen=%q; elde edilen=%q", beklenenAlanYoneticisi, eldeEdilen)
			}
			return nil
		},
		Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
			geriCagir()
			out := &client.UpdateOptions{}
			for _, f := range opts {
				f.ApplyToUpdate(out)
			}
			if eldeEdilen := out.FieldManager; beklenenAlanYoneticisi != eldeEdilen {
				t.Fatalf("yanlış alan yöneticisi: beklenen=%q; elde edilen=%q", beklenenAlanYoneticisi, eldeEdilen)
			}
			return nil
		},
		Patch: func(ctx context.Context, c client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
			geriCagir()
			out := &client.PatchOptions{}
			for _, f := range opts {
				f.ApplyToPatch(out)
			}
			if eldeEdilen := out.FieldManager; beklenenAlanYoneticisi != eldeEdilen {
				t.Fatalf("yanlış alan yöneticisi: beklenen=%q; elde edilen=%q", beklenenAlanYoneticisi, eldeEdilen)
			}
			return nil
		},
		SubResourceCreate: func(ctx context.Context, c client.Client, subResourceName string, obj client.Object, subResource client.Object, opts ...client.SubResourceCreateOption) error {
			geriCagir()
			out := &client.SubResourceCreateOptions{}
			for _, f := range opts {
				f.ApplyToSubResourceCreate(out)
			}
			if eldeEdilen := out.FieldManager; beklenenAlanYoneticisi != eldeEdilen {
				t.Fatalf("yanlış alan yöneticisi: beklenen=%q; elde edilen=%q", beklenenAlanYoneticisi, eldeEdilen)
			}
			return nil
		},
		SubResourceUpdate: func(ctx context.Context, c client.Client, subResourceName string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
			geriCagir()
			out := &client.SubResourceUpdateOptions{}
			for _, f := range opts {
				f.ApplyToSubResourceUpdate(out)
			}
			if eldeEdilen := out.FieldManager; beklenenAlanYoneticisi != eldeEdilen {
				t.Fatalf("yanlış alan yöneticisi: beklenen=%q; elde edilen=%q", beklenenAlanYoneticisi, eldeEdilen)
			}
			return nil
		},
		SubResourcePatch: func(ctx context.Context, c client.Client, subResourceName string, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
			geriCagir()
			out := &client.SubResourcePatchOptions{}
			for _, f := range opts {
				f.ApplyToSubResourcePatch(out)
			}
			if eldeEdilen := out.FieldManager; beklenenAlanYoneticisi != eldeEdilen {
				t.Fatalf("yanlış alan yöneticisi: beklenen=%q; elde edilen=%q", beklenenAlanYoneticisi, eldeEdilen)
			}
			return nil
		},
	}).Build()
}
