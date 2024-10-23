// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 The Kubernetes Authors

package versions

import (
	"fmt"
	"regexp"
)

// Platform, işletim sistemi ve mimari bilgilerini içerir.
// Her ikisi de joker karakter (*) içerebilir.
type Platform struct {
	OS   string
	Arch string
}

// Matches, bu platformun diğer platformla eşleşip eşleşmediğini belirtir,
// potansiyel olarak joker değerlerle.
func (p Platform) Matches(other Platform) bool {
	return (p.OS == other.OS || p.OS == "*" || other.OS == "*") &&
		(p.Arch == other.Arch || p.Arch == "*" || other.Arch == "*")
}

// IsWildcard, OS veya Arch'ın joker değerlere ayarlanıp ayarlanmadığını kontrol eder.
func (p Platform) IsWildcard() bool {
	return p.OS == "*" || p.Arch == "*"
}

func (p Platform) String() string {
	return fmt.Sprintf("%s/%s", p.OS, p.Arch)
}

// BaseName, belirli bir sürüm ve platformu tam olarak tanımlayan
// temel dizin adını döndürür.
func (p Platform) BaseName(ver Concrete) string {
	return fmt.Sprintf("%d.%d.%d-%s-%s", ver.Major, ver.Minor, ver.Patch, p.OS, p.Arch)
}

// ArchiveName, bu sürüm ve platform için tam arşiv adını döndürür.
func (p Platform) ArchiveName(ver Concrete) string {
	return "envtest-v" + p.BaseName(ver) + ".tar.gz"
}

// PlatformItem, bir platformu ve indirme için
// bilinen meta verilerini temsil eder.
type PlatformItem struct {
	Platform

	*Hash
}

// Hash, envtest ikili dosyalarının bir arşivinin hash'idir.
type Hash struct {
	// Hash türü.
	// controller-tools SHA512HashType kullanır.
	Type HashType

	// Hash değerinin kodlaması.
	// controller-tools HexHashEncoding kullanır.
	Encoding HashEncoding

	// Hash değeri.
	Value string
}

// HashType, bir hash türüdür.
type HashType string

const (
	// SHA512HashType, sha512 hash'ini temsil eder.
	SHA512HashType HashType = "sha512"

	// MD5HashType, md5 hash'ini temsil eder.
	MD5HashType HashType = "md5"
)

// HashEncoding, bir hash'in kodlamasıdır.
type HashEncoding string

const (
	// Base64HashEncoding, base64 kodlamasını temsil eder.
	Base64HashEncoding HashEncoding = "base64"

	// HexHashEncoding, hex kodlamasını temsil eder.
	HexHashEncoding HashEncoding = "hex"
)

// Set, belirli bir sürüm ve bu sürümün mevcut olduğu tüm platformları içerir.
type Set struct {
	Version   Concrete
	Platforms []PlatformItem
}

// ExtractWithPlatform, verilen düzenli ifade ve eşleşmesi gereken
// string'den bir sürüm ve platform çıkarır. Eşleşme bulunamazsa, Version nil olur.
//
// Düzenli ifade aşağıdaki yakalama gruplarına sahip olmalıdır:
// major, minor, patch, prelabel, prenum, os, arch ve joker sürümleri desteklememelidir.
func ExtractWithPlatform(re *regexp.Regexp, name string) (*Concrete, Platform) {
	match := re.FindStringSubmatch(name)
	if match == nil {
		return nil, Platform{}
	}
	verInfo := PatchSelectorFromMatch(match, re)
	if verInfo.AsConcrete() == nil {
		panic(fmt.Sprintf("%v", verInfo))
	}
	// RE'de joker karakterleri dışladık, bu yüzden güvenli bir şekilde dönüştürebiliriz
	return verInfo.AsConcrete(), Platform{
		OS:   match[re.SubexpIndex("os")],
		Arch: match[re.SubexpIndex("arch")],
	}
}

var (
	versionPlatformREBase = ConcreteVersionRE.String() + `-(?P<os>\w+)-(?P<arch>\w+)`
	// VersionPlatformRE, belirli sürüm-platform string'lerini eşleştirir.
	VersionPlatformRE = regexp.MustCompile(`^` + versionPlatformREBase + `$`)
	// ArchiveRE, belirli sürüm-platform.tar.gz string'lerini eşleştirir.
	// controller-tools tarafından GitHub sürümlerine yayınlanan arşivler "envtest-v" önekini kullanır (örn. "envtest-v1.30.0-darwin-amd64.tar.gz").
	ArchiveRE = regexp.MustCompile(`^envtest-v` + versionPlatformREBase + `\.tar\.gz$`)
)
