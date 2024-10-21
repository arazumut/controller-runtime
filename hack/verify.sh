#!/usr/bin/env bash

#  2018 Kubernetes Yazarları.
#
#  Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
#  bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
#  Lisans'ın bir kopyasını aşağıdaki adreste bulabilirsiniz:
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Geçerli yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
#  Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
#  HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
#  Lisans kapsamında izin verilen belirli dil kapsamındaki
#  haklar ve sınırlamalar için Lisans'a bakınız.

set -e

source $(dirname ${BASH_SOURCE})/common.sh

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
cd "${REPO_ROOT}"

header_text "modüller çalıştırılıyor"
make modules

# Sadece CI'da verify-modules çalıştır, aksi takdirde
# go modülünü yerel olarak güncellemek (geçerli bir işlem olan) `make test`'in başarısız olmasına neden olur.
if [[ -n ${CI} ]]; then
    header_text "modüller doğrulanıyor"
    make verify-modules
fi

header_text "golangci-lint çalıştırılıyor"
make lint
