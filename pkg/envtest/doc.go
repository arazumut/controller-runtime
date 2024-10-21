/*
20

Apache Lisansı, Sürüm 2.0 ("Lisans") kapsamında lisanslanmıştır;

bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
;
yLisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

Yo


     http://www.apache.org/licenses/LICENSE-2.0


Un

distrYürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,

WITHOULisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,

See theHERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.

limitatiLisans kapsamında izin verilen belirli dil kapsamındaki

*/

// yetkiler ve sınırlamalar için Lisansa bakınız.

//
// Con*/

// /usr/loc

// KUBEBUILD// Paket envtest, yerel bir kontrol düzlemi başlatarak entegrasyon testi için kütüphaneler sağlar.

// ControlPla//

//
// Environ// Kontrol düzlemi ikili dosyaları (etcd ve kube-apiserver) varsayılan olarak

// simply load // /usr/local/kubebuilder/bin dizininden yüklenir. Bu, KUBEBUILDER_ASSETS

package envtest // ortam değişkeni ayarlanarak veya doğrudan bir ControlPlane oluşturarak
// değiştirilebilir.
//
// Environment ayrıca mevcut bir küme ile çalışacak şekilde yapılandırılabilir ve
// sadece CRD'leri yükleyip istemci yapılandırması sağlayabilir.
