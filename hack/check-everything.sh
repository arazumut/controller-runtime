#!/usr/bin/env bash

#  Copyright 2018 The Kubernetes Authors.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

# Hata durumunda çık
set -o errexit
# Tanımsız değişken kullanımı hatası
set -o nounset
# Pipe hatalarını yakala
set -o pipefail

# Hack dizinini belirle
hack_dir=$(dirname "${BASH_SOURCE[0]}")
source "${hack_dir}/common.sh"

# Geçici dizinler
tmp_root=/tmp
kb_root_dir=$tmp_root/kubebuilder

# Go araç zinciri sürümünü ayarla
export GOTOOLCHAIN="go$(make go-version)"

# Doğrulama scriptlerini çalıştır
"${hack_dir}/verify.sh"

# Envtest sürümü
ENVTEST_K8S_VERSION=${ENVTEST_K8S_VERSION:-"1.28.0"}

# Envtest araçlarını kur
header_text "Gerekirse setup-envtest ile envtest araçları@${ENVTEST_K8S_VERSION} kuruluyor"
tmp_bin=/tmp/cr-tests-bin
(
    # Kullanıcı için kurulum yapma varsayımı yapma
    cd "${hack_dir}/../tools/setup-envtest"
    GOBIN=${tmp_bin} go install .
)
export KUBEBUILDER_ASSETS="$(${tmp_bin}/setup-envtest use --use-env -p path "${ENVTEST_K8S_VERSION}")"

# Testleri çalıştır
"${hack_dir}/test-all.sh"

# Örneklerin derlendiğini doğrula (go install ile)
header_text "Örneklerin derlendiğini doğrulama (go install ile)"
go install ${MOD_OPT} ./examples/builtins
go install ${MOD_OPT} ./examples/crd

echo "başarılı"
exit 0
