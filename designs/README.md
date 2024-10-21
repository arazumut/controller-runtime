Designs
=======

Bunlar Controller Runtime'daki değişiklikler için tasarım belgeleridir. Bu belgeler, Controller Runtime yazma sürecine giren tasarım süreçlerini belgelemeye yardımcı olmak için vardır, ancak güncel olmayabilirler (aşağıya bakın).

Controller Runtime'daki tüm değişikliklerin bir tasarım belgesine ihtiyacı yoktur - sadece büyük olanların. En iyi yargınızı kullanın.

Bir tasarım belgesi gönderirken, bir kavram kanıtı yazmış olmanızı teşvik ederiz ve kavram kanıtı PR'ını tasarım belgesi ile eşzamanlı olarak göndermek tamamen kabul edilebilir, çünkü kavram kanıtı süreci kırışıklıkları gidermeye yardımcı olabilir ve şablonun `Örnek` bölümüne yardımcı olabilir.

## Güncel Olmayan Tasarımlar

**Controller Runtime belgeleri
[GoDoc](https://pkg.go.dev/sigs.k8s.io/controller-runtime) Controller Runtime için kanonik, güncel referans ve mimari dokümantasyon olarak kabul edilmelidir.**

Ancak, güncel olmayan bir tasarım belgesi görürseniz, onu böyle işaretleyen bir PR göndermekten çekinmeyin ve neden değişikliklerin yapıldığını belgeleyen sorunlara bağlantı ekleyin. Örneğin:

```markdown

# Güncel Değil

Bu değişiklik güncel değildir. Süslü parantezlerin yazılması sinir bozucu olduğu için işlevleri tamamen terk etmek zorunda kaldık ve kullanıcıların özel işlevselliği Common LISP dizeleri kullanarak belirtmelerini sağladık. Daha fazla bilgi için #000'ye bakın.
```
