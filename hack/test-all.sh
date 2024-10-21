#!/usr/bin/env bash

#  2018 Kubernetes Yazarları.
#
#  Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
#  bu dosyayı ancak Lisans uyarınca kullanabilirsiniz.
#  Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Geçerli yasa veya yazılı izin gereği aksi belirtilmedikçe,
#  Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
#  HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
#  Lisans kapsamında izin verilen belirli dil kapsamındaki
#  haklar ve sınırlamalar için Lisansa bakınız.

set -e

source $(dirname ${BASH_SOURCE[0]})/common.sh

header_text "go test çalıştırılıyor"

if [[ -n ${ARTIFACTS:-} ]]; then
  GINKGO_ARGS="-ginkgo.junit-report=junit-report.xml"
fi

result=0
go test -v -race ${P_FLAG} ${MOD_OPT} ./... --ginkgo.fail-fast ${GINKGO_ARGS} || result=$?

if [[ -n ${ARTIFACTS:-} ]]; then
  mkdir -p ${ARTIFACTS}
  for file in $(find . -name "*junit-report.xml"); do
    new_file=${file#./}
    new_file=${new_file%/junit-report.xml}
    new_file=${new_file//"/"/"-"}
    mv "$file" "$ARTIFACTS/junit_${new_file}.xml"
  done
fi

exit $result
