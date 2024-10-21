# FAQ

### S: Bir denetleyicinin hangi tür nesneye referans verdiğini nasıl anlarım?

**C**: Her denetleyici yalnızca bir nesne türünü uzlaştırmalıdır. Diğer etkilenen nesneler, `handler.EnqueueRequestForOwner` veya `handler.EnqueueRequestsFromMapFunc` olay işleyicilerini ve potansiyel olarak dizinleri kullanarak tek bir kök nesne türüne eşlenmelidir. Ardından, Reconcile metodunuz, verilen kök nesnelerin *tüm* durumunu uzlaştırmaya çalışmalıdır.

### S: Farklı olay türleri (örneğin, oluşturma, güncelleme, silme) için reconciler'ımda farklı mantık nasıl kullanabilirim?

**C**: Kullanılmamalıdır. Reconcile fonksiyonları idempotent olmalı ve ihtiyaç duyduğu tüm durumu okuyarak, ardından güncellemeleri yazarak her zaman durumu uzlaştırmalıdır. Bu, reconciler'ınızın genel olaylara doğru şekilde yanıt vermesini, atlanan veya birleştirilen olaylara uyum sağlamasını ve uygulama başlatma işlemiyle kolayca başa çıkmasını sağlar. Denetleyici, bir eşleme değişirse hem eski hem de yeni nesneler için reconcile isteklerini sıraya alacaktır, ancak artık referans edilmeyen durumu temizlemek için yeterli bilgiye sahip olduğunuzdan emin olmak sizin sorumluluğunuzdadır.

### S: Önbellekten okursam önbelleğim eski olabilir! Bununla nasıl başa çıkmalıyım?

**C**: Durumunuza bağlı olarak alınabilecek birkaç farklı yaklaşım vardır.

- Mümkün olduğunda iyimser kilitlemeden yararlanın: Oluşturduğunuz nesneler için belirleyici adlar kullanın, böylece Kubernetes API sunucusu nesnenin zaten var olup olmadığını size bildirecektir. Kubernetes'teki birçok denetleyici bu yaklaşımı benimser: StatefulSet denetleyicisi oluşturduğu her pod'a belirli bir sayı eklerken, Deployment denetleyicisi pod şablon spesifikasyonunu hash'ler ve bunu ekler.

- Belirleyici adlardan yararlanamadığınız birkaç durumda (örneğin, generateName kullanırken), hangi eylemleri gerçekleştirdiğinizi izlemek ve belirli bir süre sonra gerçekleşmezlerse tekrarlanmaları gerektiğini varsaymak faydalı olabilir (örneğin, bir yeniden sıraya alma sonucu kullanarak). Bu, ReplicaSet denetleyicisinin yaptığı şeydir.

Genel olarak, denetleyicinizi bilgilerin sonunda doğru olacağı, ancak biraz eski olabileceği varsayımıyla yazın. Reconcile fonksiyonunuzun her çalıştığında dünyanın tüm durumunu uyguladığından emin olun. Bunların hiçbiri sizin için işe yaramazsa, doğrudan API sunucusundan okuyan bir istemci oluşturabilirsiniz, ancak bu genellikle son çare olarak kabul edilir ve yukarıdaki iki yaklaşım genellikle çoğu durumu kapsamalıdır.

### S: Sahte istemci nerede? Nasıl kullanırım?

**C**: Sahte istemci [mevcut](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client/fake), ancak genellikle gerçek bir API sunucusuna karşı test yapmak için [envtest.Environment](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest#Environment) kullanmanızı öneririz. Deneyimlerimize göre, sahte istemciler kullanan testler, gerçek bir API sunucusunun kötü yazılmış izlenimlerini kademeli olarak yeniden uygular, bu da bakımı zor, karmaşık test kodlarına yol açar.

### S: Testleri nasıl yazmalıyım? Başlamak için herhangi bir öneri var mı?

- Gerçek bir API sunucusunu başlatmak için yukarıda belirtilen [envtest.Environment](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest#Environment) kullanın, birini taklit etmeye çalışmak yerine.

- Testlerinizi, Kubernetes API'leri ile çalışırken belirli bir API çağrısı setinin yapıldığını değil, dünyanın durumunun beklediğiniz gibi olup olmadığını kontrol edecek şekilde yapılandırın. Bu, denetleyicilerinizin iç işleyişini değiştirmeden ve testlerinizi değiştirmeden daha kolay yeniden düzenlemenizi ve geliştirmenizi sağlar.

- API sunucusuyla etkileşimde bulunduğunuz her zaman, değişikliklerin yazma zamanı ile reconcile zamanı arasında biraz gecikme olabileceğini unutmayın.

### S: Bir tür için kayıtlı Kind yok hataları nedir?

**C**: Muhtemelen tam olarak ayarlanmış bir Scheme eksik. Scheme'ler, Kubernetes'teki Go türleri ile grup-sürüm-türler arasındaki eşlemeyi kaydeder. Genel olarak, uygulamanızın ihtiyaç duyduğu API gruplarından (Kubernetes türleri veya kendi türleriniz olsun) türleri içeren kendi Scheme'ine sahip olmalıdır. Daha fazla bilgi için [scheme builder belgelerine](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/scheme) bakın.

