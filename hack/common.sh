#!/usr/bin/env bash

#  2018 Kubernetes Yazarları tarafından oluşturulmuştur.
#
#  Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
#  bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
#  Lisansın bir kopyasını aşağıdaki adresten edinebilirsiniz:
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Geçerli yasa veya yazılı izin gerektirmedikçe,
#  Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
#  herhangi bir garanti veya koşul olmaksızın, açık veya zımni.
#  Lisans kapsamındaki belirli dil izinleri ve
#  sınırlamaları için Lisans'a bakınız.

set -e

# Bu betikte izlemeyi etkinleştirmek için TRACE değişkenini
# ortamınızda herhangi bir değere ayarlayın:
#
# $ TRACE=1 test.sh
TRACE=${TRACE:-""}
if [ -n "$TRACE" ]; then
  set -x
fi

# Modüllerin etkin olup olmadığını kontrol et
(go mod edit -json &>/dev/null)
MODULES_ENABLED=$?

MOD_OPT=""
MODULES_OPT=${MODULES_OPT:-""}
if [[ -n "${MODULES_OPT}" && $MODULES_ENABLED -eq 0 ]]; then
    MOD_OPT="-mod=${MODULES_OPT}"
fi

# Bu betikte renkleri kapatmak için NO_COLOR değişkenini
# ortamınızda herhangi bir değere ayarlayın:
#
# $ NO_COLOR=1 test.sh
NO_COLOR=${NO_COLOR:-""}
if [ -z "$NO_COLOR" ]; then
  header=$'\e[1;33m'
  reset=$'\e[0m'
else
  header=''
  reset=''
fi

function header_text {
  echo "$header$*$reset"
}
