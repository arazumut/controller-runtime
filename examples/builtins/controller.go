/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa veya yazılı izin gereği olmadıkça,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
haklar ve sınırlamalar için Lisansa bakınız.
*/

package main

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// reconcileReplicaSet ReplicaSet'leri uzlaştırır
type reconcileReplicaSet struct {
	// client, APIServer'dan nesneleri almak için kullanılabilir.
	client client.Client
}

// reconcile.Reconciler'ı uygula ki kontrolcü nesneleri uzlaştırabilsin
var _ reconcile.Reconciler = &reconcileReplicaSet{}

func (r *reconcileReplicaSet) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	// request'i tekrar tekrar yazmamak için uygun bir log nesnesi oluştur
	log := log.FromContext(ctx)

	// ReplicaSet'i önbellekten al
	rs := &appsv1.ReplicaSet{}
	err := r.client.Get(ctx, request.NamespacedName, rs)
	if errors.IsNotFound(err) {
		log.Error(nil, "ReplicaSet bulunamadı")
		return reconcile.Result{}, nil
	}

	if err != nil {
		return reconcile.Result{}, fmt.Errorf("ReplicaSet alınamadı: %+v", err)
	}

	// Etiketin ayarlanıp ayarlanmadığını kontrol et ve ayarlanmamışsa ReplicaSet'i güncelle
	if rs.Labels == nil {
		rs.Labels = map[string]string{}
	}
	if rs.Labels["hello"] == "world" {
		return reconcile.Result{}, nil
	}

	// ReplicaSet'i yazdır
	log.Info("ReplicaSet uzlaştırılıyor", "konteyner adı", rs.Spec.Template.Spec.Containers[0].Name)

	// Etiket eksikse ayarla
	rs.Labels["hello"] = "world"

	// ReplicaSet'i güncelle
	err = r.client.Update(ctx, rs)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("ReplicaSet yazılamadı: %+v", err)
	}

	return reconcile.Result{}, nil
}
