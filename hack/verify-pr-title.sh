#!/bin/bash

# Copyright 2024 The Kubernetes Authors.
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

# Regex desenlerini tanımla
WIP_REGEX="^\W?WIP\W"
TAG_REGEX="^\[[[:alnum:]\._-]*\]"
PR_TITLE="$1"

# Başlıktan WIP ve etiketleri kaldır
trimmed_title=$(echo "$PR_TITLE" | sed -E "s/$WIP_REGEX//" | sed -E "s/$TAG_REGEX//" | xargs)

# Yaygın emojileri metin formundan gerçek emojilere dönüştür
trimmed_title=$(echo "$trimmed_title" | sed -E "s/:warning:/⚠/g")
trimmed_title=$(echo "$trimmed_title" | sed -E "s/:sparkles:/✨/g")
trimmed_title=$(echo "$trimmed_title" | sed -E "s/:bug:/🐛/g")
trimmed_title=$(echo "$trimmed_title" | sed -E "s/:book:/📖/g")
trimmed_title=$(echo "$trimmed_title" | sed -E "s/:rocket:/🚀/g")
trimmed_title=$(echo "$trimmed_title" | sed -E "s/:seedling:/🌱/g")

# PR türü öneki kontrol et
if [[ "$trimmed_title" =~ ^(⚠|✨|🐛|📖|🚀|🌱) ]]; then
    echo "PR başlığı geçerli: $trimmed_title"
else
    echo "Hata: Başlıkta eşleşen bir PR türü göstergesi bulunamadı."
    echo "PR başlığınızın şu öneklerden birine sahip olması gerekiyor:"
    echo "- Kırıcı değişiklik: ⚠ (:warning:)"
    echo "- Kırıcı olmayan özellik: ✨ (:sparkles:)"
    echo "- Yama düzeltmesi: 🐛 (:bug:)"
    echo "- Dokümantasyon: 📖 (:book:)"
    echo "- Sürüm: 🚀 (:rocket:)"
    echo "- Altyapı/Testler/Diğer: 🌱 (:seedling:)"
    exit 1
fi

# PR başlığının Issue veya PR numarası içermediğini kontrol et
if [[ "$trimmed_title" =~ \#[0-9]+ ]]; then
    echo "Hata: PR başlığı issue veya PR numarası içermemelidir."
    echo "Issue numaraları PR gövdesinde \"Fixes #XYZ\" (eğer issue veya PR'ı kapatıyorsa) veya \"Related to #XYZ\" (eğer sadece ilgiliyse) şeklinde yer almalıdır."
    exit 1
fi
