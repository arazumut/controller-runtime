/*
2019 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN.
Lisans kapsamındaki izinleri ve sınırlamaları belirten
Lisans'a bakınız.
*/

package builder_test

import (
	"os"

	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	examplegroup "sigs.k8s.io/controller-runtime/examples/crd/pkg"
)

// Bu örnek, CRD ChaosPod için bir yönetici tarafından yönetilen basit bir webhook oluşturmak için webhook builder kullanır.
// Ardından yöneticiyi başlatır.
func ExampleWebhookBuilder() {
	var log = logf.Log.WithName("webhookbuilder-örnek")

	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		log.Error(err, "yönetici oluşturulamadı")
		os.Exit(1)
	}

	err = builder.
		WebhookManagedBy(mgr).         // WebhookManagedBy oluştur
		For(&examplegroup.ChaosPod{}). // ChaosPod bir CRD'dir.
		Complete()
	if err != nil {
		log.Error(err, "webhook oluşturulamadı")
		os.Exit(1)
	}

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "yönetici başlatılamadı")
		os.Exit(1)
	}
}
