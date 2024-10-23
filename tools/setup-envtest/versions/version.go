// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 The Kubernetes Authors

package versions

import (
	"fmt"
	"strconv"
)

// NB(directxman12): Bu kodun çoğu özel olarak yazılmıştır çünkü
// a) Standart kütüphanelerin hiçbiri hashlenebilir sürüm türlerine sahip değil (geçerli nedenlerle,
//    ancak kullanım durumumuz için sınırlı bir alt küme kullanabiliriz)
// b) Herkesin seçicilerin nasıl çalıştığına dair kendi tanımı var

// NB(directxman12): Ön sürüm desteği seçicilerle karmaşıktır
// Eğer buna ihtiyacımız olursa, joker karakterli bir ön sürüm türünün ne anlama geldiğini dikkatlice düşünün
// ("ön sürüm değil" dahil mi?), ve <=1.17.3-x.x ne anlama geliyor.

// Concrete, Kubernetes tarzı semver sürümünün somut bir halidir.
type Concrete struct {
	Major, Minor, Patch int
}

// AsConcrete bu sürümü döndürür.
func (c Concrete) AsConcrete() *Concrete {
	return &c
}

// NewerThan verilen diğer sürümün bu sürümden daha yeni olup olmadığını kontrol eder.
func (c Concrete) NewerThan(other Concrete) bool {
	if c.Major != other.Major {
		return c.Major > other.Major
	}
	if c.Minor != other.Minor {
		return c.Minor > other.Minor
	}
	return c.Patch > other.Patch
}

// Matches bu sürümün diğerine eşit olup olmadığını kontrol eder.
func (c Concrete) Matches(other Concrete) bool {
	return c == other
}

func (c Concrete) String() string {
	return fmt.Sprintf("%d.%d.%d", c.Major, c.Minor, c.Patch)
}

// PatchSelector, yamanın joker karakter olduğu bir dizi sürümü seçer.
type PatchSelector struct {
	Major, Minor int
	Patch        PointVersion
}

func (s PatchSelector) String() string {
	return fmt.Sprintf("%d.%d.%s", s.Major, s.Minor, s.Patch)
}

// Matches verilen sürümün bu seçiciyle eşleşip eşleşmediğini kontrol eder.
func (s PatchSelector) Matches(ver Concrete) bool {
	return s.Major == ver.Major && s.Minor == ver.Minor && s.Patch.Matches(ver.Patch)
}

// AsConcrete bu seçicide joker karakterler varsa nil döndürür,
// ve aksi takdirde bu seçicinin seçtiği somut sürümü döndürür.
func (s PatchSelector) AsConcrete() *Concrete {
	if s.Patch == AnyPoint {
		return nil
	}

	return &Concrete{
		Major: s.Major,
		Minor: s.Minor,
		Patch: int(s.Patch), // joker karakterleri yukarıda kontrol ettik, bu yüzden cast etmek güvenli
	}
}

// TildeSelector [X.Y.Z, X.Y+1.0) aralığını seçer.
type TildeSelector struct {
	Concrete
}

// Matches verilen sürümün bu seçiciyle eşleşip eşleşmediğini kontrol eder.
func (s TildeSelector) Matches(ver Concrete) bool {
	if s.Concrete.Matches(ver) {
		// kolay, "tam" eşleşme
		return true
	}
	return ver.Major == s.Major && ver.Minor == s.Minor && ver.Patch >= s.Patch
}

func (s TildeSelector) String() string {
	return "~" + s.Concrete.String()
}

// AsConcrete nil döndürür (bu asla somut bir sürüm değildir).
func (s TildeSelector) AsConcrete() *Concrete {
	return nil
}

// LessThanSelector verilen sürümden daha eski sürümleri seçer
// (özellikle temizleme için kullanışlıdır).
type LessThanSelector struct {
	PatchSelector
	OrEquals bool
}

// Matches verilen sürümün bu seçiciyle eşleşip eşleşmediğini kontrol eder.
func (s LessThanSelector) Matches(ver Concrete) bool {
	if s.Major != ver.Major {
		return s.Major > ver.Major
	}
	if s.Minor != ver.Minor {
		return s.Minor > ver.Minor
	}
	if !s.Patch.Matches(ver.Patch) {
		// eşleşme kuralları joker karakteri dışlar, bu yüzden normal sayılar olarak karşılaştırmak sorun değil
		return int(s.Patch) > ver.Patch
	}
	return s.OrEquals
}

func (s LessThanSelector) String() string {
	if s.OrEquals {
		return "<=" + s.PatchSelector.String()
	}
	return "<" + s.PatchSelector.String()
}

// AsConcrete nil döndürür (bu asla somut bir sürüm değildir).
func (s LessThanSelector) AsConcrete() *Concrete {
	return nil
}

// AnySelector herhangi bir sürümle eşleşir.
type AnySelector struct{}

// Matches verilen sürümün bu seçiciyle eşleşip eşleşmediğini kontrol eder.
func (AnySelector) Matches(_ Concrete) bool { return true }

// AsConcrete nil döndürür (bu asla somut bir sürüm değildir).
func (AnySelector) AsConcrete() *Concrete { return nil }
func (AnySelector) String() string        { return "*" }

// Selector somut bir sürümü veya sürüm aralığını seçer.
type Selector interface {
	// AsConcrete bu seçiciyi somut bir sürüm olarak döndürmeye çalışır.
	// Seçici yalnızca tek bir sürümle eşleşiyorsa,
	// onu döndürür, aksi takdirde nil döner.
	AsConcrete() *Concrete
	// Matches bu seçicinin verilen somut sürümle eşleşip eşleşmediğini kontrol eder.
	Matches(ver Concrete) bool
	String() string
}

// Spec bazı sürümlerle veya sürüm aralıklarıyla eşleşir ve
// bir sürüm seçerken yerel ve uzak sunucuyla nasıl başa çıkacağımızı söyler.
type Spec struct {
	Selector

	// CheckLatest, seçicimizle eşleşen en son sürümü bulmak için
	// uzak sunucuyu kontrol etmemizi söyler, sadece yerel sürümlerle
	// yetinmek yerine.
	CheckLatest bool
}

// MakeConcrete bu spec'in içeriğini verilen somut sürümle
// eşleşecek şekilde değiştirir (sunucudan en son sürümü kontrol etmeden).
func (s *Spec) MakeConcrete(ver Concrete) {
	s.Selector = ver
	s.CheckLatest = false
}

// AsConcrete alttaki seçiciyi somut bir sürüm olarak döndürür, eğer
// mümkünse.
func (s Spec) AsConcrete() *Concrete {
	return s.Selector.AsConcrete()
}

// Matches alttaki seçicinin verilen sürümle eşleşip eşleşmediğini kontrol eder.
func (s Spec) Matches(ver Concrete) bool {
	return s.Selector.Matches(ver)
}

func (s Spec) String() string {
	res := s.Selector.String()
	if s.CheckLatest {
		res += "!"
	}
	return res
}

// PointVersion joker karakter (yama) sürümünü veya somut sayıyı temsil eder.
type PointVersion int

const (
	// AnyPoint herhangi bir nokta sürümüyle eşleşir.
	AnyPoint PointVersion = -1
)

// Matches bir nokta sürümünün somut bir nokta sürümüyle uyumlu olup olmadığını kontrol eder.
// İki nokta sürümü uyumludur eğer
// a) her ikisi de somutsa
// b) biri joker karakterse.
func (p PointVersion) Matches(other int) bool {
	switch p {
	case AnyPoint:
		return true
	default:
		return int(p) == other
	}
}

func (p PointVersion) String() string {
	switch p {
	case AnyPoint:
		return "*"
	default:
		return strconv.Itoa(int(p))
	}
}

var (
	// LatestVersion uzak sunucudaki en son sürümle eşleşir.
	LatestVersion = Spec{
		Selector:    AnySelector{},
		CheckLatest: true,
	}
	// AnyVersion herhangi bir yerel veya uzak sürümle eşleşir.
	AnyVersion = Spec{
		Selector: AnySelector{},
	}
)
