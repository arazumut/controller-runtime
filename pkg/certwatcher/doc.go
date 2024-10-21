/*
2021 Kubernetes Yazarları tarafından telif hakkı saklıdır.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisansa uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:

    http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni olarak.
Lisans altındaki izinleri ve sınırlamaları yöneten özel dil için
Lisans'a bakınız.
*/

/*
certwatcher paketi, tls sunucularında kullanılmak üzere diskten Sertifikaları yeniden yüklemek için bir yardımcıdır.
`tls.Config`'den çağrılabilecek ve tls.Listener'ınıza geçirilebilecek bir yardımcı fonksiyon `GetCertificate` sağlar.
Ayrıntılı bir örnek sunucu için pkg/webhook/server.go dosyasına bakın.
*/
package certwatcher
