#!/usr/bin/env bash

#  2018 Kubernetes Yazarları.
#
#  Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
#  bu dosyayı ancak Lisansa uygun şekilde kullanabilirsiniz.
#  Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Geçerli yasa veya yazılı izin gerektirmedikçe,
#  Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
#  herhangi bir garanti veya koşul olmaksızın,
#  açık veya zımni olarak. Lisans kapsamındaki izinleri ve
#  sınırlamaları yöneten özel dil için Lisansa bakınız.

set -e

# Hata ayıklama için izleme modunu etkinleştir
export TRACE=1

# Prow'da varsayılan olarak dahil edilmemiş veya mevcut değil
export PATH=$(go env GOPATH)/bin:$PATH
mkdir -p $(go env GOPATH)/bin

# check-everything.sh dosyasını çalıştır
$(dirname "${BASH_SOURCE[0]}")/check-everything.sh
