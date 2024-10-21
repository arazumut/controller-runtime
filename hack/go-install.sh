#!/usr/bin/env bash
# Copyright 2021 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

# İlk parametre olarak modül sağlanmalıdır
if [ -z "${1:-}" ]; then
  echo "İlk parametre olarak modül sağlanmalıdır"
  exit 1
fi

# İkinci parametre olarak binary adı sağlanmalıdır
if [ -z "${2:-}" ]; then
  echo "İkinci parametre olarak binary adı sağlanmalıdır"
  exit 1
fi

# Üçüncü parametre olarak versiyon sağlanmalıdır
if [ -z "${3:-}" ]; then
  echo "Üçüncü parametre olarak versiyon sağlanmalıdır"
  exit 1
fi

# GOBIN değişkeni ayarlanmış olmalıdır
if [ -z "${GOBIN:-}" ]; then
  echo "GOBIN ayarlanmamış. Bin dosyasını belirli bir dizine kurmak için GOBIN ayarlanmalıdır."
  exit 1
fi

# Eski binary dosyasını sil
rm -f "${GOBIN}/${2}"* || true

# Belirtilen golang modülünü kur
go install "${1}@${3}"
mv "${GOBIN}/${2}" "${GOBIN}/${2}-${3}"
ln -sf "${GOBIN}/${2}-${3}" "${GOBIN}/${2}"
