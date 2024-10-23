// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 The Kubernetes Authors

package versions

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	// baseVersionRE, X.Y.Z, X.Y veya X.Y.{*|x} formatında semver-benzeri bir versiyondur.
	baseVersionRE = `(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)(?:\.(?P<patch>0|[1-9]\d*|x|\*))?`
	// versionExprRE, FromExpr için geçerli versiyon girişlerini eşleştirir.
	versionExprRE = regexp.MustCompile(`^(?P<sel><|~|<=)?` + baseVersionRE + `(?P<latest>!)?$`)

	// ConcreteVersionRE, bir string içinde herhangi bir yerde somut bir versiyonu eşleştirir.
	ConcreteVersionRE = regexp.MustCompile(`(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)`)
)

// FromExpr, bir stringden semver formatında bir versiyon çıkarır.
// X, Y ve Z joker karakterler ('*', 'x') olabilir ve ön sürüm isimleri ve numaraları da joker karakterler olabilir.
// Tüm string, aşağıdaki gibi bir versiyon seçici olarak kabul edilir:
//   - X.Y.Z, x, y ve z >= 0 olan int'lerdir ve Z '*' veya 'x' olabilir
//   - X.Y, X.Y.* ile eşdeğerdir
//   - ~X.Y.Z, >= X.Y.Z && < X.Y+1.0 anlamına gelir
//   - <X.Y.Z, X.Y.Z'den daha eski anlamına gelir (temizlik için kullanışlıdır, <= için de benzer şekilde)
//   - Sonunda '!' işareti varsa, en son sürümleri yerel eşleşmeler yerine API sunucusundan kontrol etmeye zorlar.
func FromExpr(expr string) (Spec, error) {
	match := versionExprRE.FindStringSubmatch(expr)
	if match == nil {
		return Spec{}, fmt.Errorf("'%q' versiyon stringi olarak parse edilemedi. "+
			"X.Y.Z formatında olmalı, burada Z '*', 'x' olabilir veya tamamen bırakılabilir. "+
			"Opsiyonel olarak ~|<|<= ile başlayabilir ve opsiyonel olarak ! ile bitebilir.", expr)
	}
	verInfo := PatchSelectorFromMatch(match, versionExprRE)
	latest := match[versionExprRE.SubexpIndex("latest")] == "!"
	sel := match[versionExprRE.SubexpIndex("sel")]
	spec := Spec{
		CheckLatest: latest,
	}
	if sel == "" {
		spec.Selector = verInfo
		return spec, nil
	}

	switch sel {
	case "<", "<=":
		spec.Selector = LessThanSelector{PatchSelector: verInfo, OrEquals: sel == "<="}
	case "~":
		// patch ve preNum >= karşılaştırmaları olduğundan, bir seçici ile joker karakterler kullanırsak
		// bunları sıfıra ayarlayabiliriz.
		if verInfo.Patch == AnyPoint {
			verInfo.Patch = PointVersion(0)
		}
		baseVer := *verInfo.AsConcrete()
		spec.Selector = TildeSelector{Concrete: baseVer}
	default:
		panic("ulaşılamaz: FromExpr ve RE'si arasında seçicide uyumsuzluk")
	}

	return spec, nil
}

// PointVersionFromValidString, string temsilinden bir nokta versiyonu çıkarır.
// Bu temsil >= 0 olan bir sayı veya x|* (AnyPoint) olabilir.
// Başka bir şey panik oluşturur (bu regexlerden çıkarılan stringler üzerinde kullanılır).
func PointVersionFromValidString(str string) PointVersion {
	switch str {
	case "*", "x":
		return AnyPoint
	default:
		ver, err := strconv.Atoi(str)
		if err != nil {
			panic(err)
		}
		return PointVersion(ver)
	}
}

// PatchSelectorFromMatch, önceden doğrulanmış bölümlerden ParseExpr kurallarına göre basit bir seçici oluşturur.
// re, major, minor, patch, prenum ve prelabel için isim yakalamalarını içermelidir.
// Herhangi bir kötü giriş panik oluşturabilir. RE eşleşmesinden alınan parçalarla kullanın.
func PatchSelectorFromMatch(match []string, re *regexp.Regexp) PatchSelector {
	// RE ile zaten parse edildi, hataları göz ardı etmek güvenli olmalı
	major, err := strconv.Atoi(match[re.SubexpIndex("major")])
	if err != nil {
		panic("patch seçici olarak geçersiz giriş (geçersiz durum) iletildi")
	}
	minor, err := strconv.Atoi(match[re.SubexpIndex("minor")])
	if err != nil {
		panic("patch seçici olarak geçersiz giriş (geçersiz durum) iletildi")
	}

	// patch isteğe bağlıdır, bırakılırsa joker karakter anlamına gelir
	patch := AnyPoint
	if patchRaw := match[re.SubexpIndex("patch")]; patchRaw != "" {
		patch = PointVersionFromValidString(patchRaw)
	}
	return PatchSelector{
		Major: major,
		Minor: minor,
		Patch: patch,
	}
}
