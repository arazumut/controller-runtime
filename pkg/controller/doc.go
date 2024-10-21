/*
2018 Kubernetes Yazarları tarafından oluşturulmuştur.

Apache Lisansı, Sürüm 2.0 ("Lisans") altında lisanslanmıştır;
bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

    http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ OLMAKSIZIN; açık veya zımni garantiler dahil ancak bunlarla sınırlı olmamak üzere.
Lisans altındaki izinleri ve sınırlamaları yöneten özel dil için Lisansa bakınız.
*/

/*
Controller paketi, Kontrolörler oluşturmak için türler ve işlevler sağlar. Kontrolörler Kubernetes API'lerini uygular.

# Oluşturma

Yeni bir Kontrolör oluşturmak için önce bir manager.Manager oluşturun ve bunu controller.New işlevine geçirin.
Kontrolör, Manager.Start çağrılarak başlatılmalıdır.
*/
package controller
