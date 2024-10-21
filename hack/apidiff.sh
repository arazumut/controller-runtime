#!/usr/bin/env bash

#  2018 Kubernetes Yazarları tarafından oluşturulmuştur.
#
#  Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
#  bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
#  Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Geçerli yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
#  yazılım Lisans kapsamında "OLDUĞU GİBİ" dağıtılır,
#  herhangi bir garanti veya koşul olmaksızın, açık veya zımni.
#  Lisans kapsamındaki izinler ve sınırlamalar hakkında daha fazla bilgi için Lisansı inceleyin.

set -o errexit  # Hata durumunda scripti durdur
set -o nounset  # Tanımsız değişken kullanımı hatası
set -o pipefail # Pipe içindeki herhangi bir komut hata verirse scripti durdur

source "$(dirname "${BASH_SOURCE[0]}")/common.sh"

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
cd "${REPO_ROOT}"

export GOTOOLCHAIN="go$(make go-version)"

header_text "API farkını doğrulama"
echo "*** go-apidiff çalıştırılıyor ***"
APIDIFF_OLD_COMMIT="${PULL_BASE_SHA}" make verify-apidiff
