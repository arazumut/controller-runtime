/*
2021 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMADAN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisansa bakınız.
*/

package versions_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "sigs.k8s.io/controller-runtime/tools/setup-envtest/versions"
)

var _ = Describe("Concrete", func() {
	It("aynı sürümle eşleşmeli", func() {
		ver16 := Concrete{Major: 1, Minor: 16}
		ver17 := Concrete{Major: 1, Minor: 17}
		Expect(ver16.Matches(ver16)).To(BeTrue(), "aynı sürümle eşleşmeli")
		Expect(ver16.Matches(ver17)).To(BeFalse(), "farklı bir sürümle eşleşmemeli")
	})
	It("X.Y.Z olarak serileştirilmeli", func() {
		Expect(Concrete{Major: 1, Minor: 16, Patch: 3}.String()).To(Equal("1.16.3"))
	})
	Describe("diğer sürümlere göre sıralama yaparken", func() {
		ver1163 := Concrete{Major: 1, Minor: 16, Patch: 3}
		Specify("daha yeni yama daha yeni olmalı", func() {
			Expect(ver1163.NewerThan(Concrete{Major: 1, Minor: 16})).To(BeTrue())
		})
		Specify("daha yeni küçük sürüm daha yeni olmalı", func() {
			Expect(ver1163.NewerThan(Concrete{Major: 1, Minor: 15, Patch: 3})).To(BeTrue())
		})
		Specify("daha yeni ana sürüm daha yeni olmalı", func() {
			Expect(ver1163.NewerThan(Concrete{Major: 0, Minor: 16, Patch: 3})).To(BeTrue())
		})
	})
})

var _ = Describe("Platform", func() {
	Specify("somut bir platform tam olarak kendisiyle eşleşmeli", func() {
		plat1 := Platform{OS: "linux", Arch: "amd64"}
		plat2 := Platform{OS: "linux", Arch: "s390x"}
		plat3 := Platform{OS: "windows", Arch: "amd64"}
		Expect(plat1.Matches(plat1)).To(BeTrue(), "kendisiyle eşleşmeli")
		Expect(plat1.Matches(plat2)).To(BeFalse(), "farklı bir mimariyi reddetmeli")
		Expect(plat1.Matches(plat3)).To(BeFalse(), "farklı bir işletim sistemini reddetmeli")
	})
	Specify("joker mimari herhangi bir mimariyle eşleşmeli", func() {
		sel := Platform{OS: "linux", Arch: "*"}
		plat1 := Platform{OS: "linux", Arch: "amd64"}
		plat2 := Platform{OS: "linux", Arch: "s390x"}
		plat3 := Platform{OS: "windows", Arch: "amd64"}
		Expect(sel.Matches(sel)).To(BeTrue(), "kendisiyle eşleşmeli")
		Expect(sel.Matches(plat1)).To(BeTrue(), "aynı işletim sistemiyle bazı mimarilerle eşleşmeli")
		Expect(sel.Matches(plat2)).To(BeTrue(), "aynı işletim sistemiyle başka bir mimariyle eşleşmeli")
		Expect(plat1.Matches(plat3)).To(BeFalse(), "farklı bir işletim sistemini reddetmeli")
	})
	Specify("joker işletim sistemi herhangi bir işletim sistemiyle eşleşmeli", func() {
		sel := Platform{OS: "*", Arch: "amd64"}
		plat1 := Platform{OS: "linux", Arch: "amd64"}
		plat2 := Platform{OS: "windows", Arch: "amd64"}
		plat3 := Platform{OS: "linux", Arch: "s390x"}
		Expect(sel.Matches(sel)).To(BeTrue(), "kendisiyle eşleşmeli")
		Expect(sel.Matches(plat1)).To(BeTrue(), "aynı mimariyle bazı işletim sistemleriyle eşleşmeli")
		Expect(sel.Matches(plat2)).To(BeTrue(), "aynı mimariyle başka bir işletim sistemiyle eşleşmeli")
		Expect(plat1.Matches(plat3)).To(BeFalse(), "farklı bir mimariyi reddetmeli")
	})
	It("joker işletim sistemini joker platform olarak rapor etmeli", func() {
		Expect(Platform{OS: "*", Arch: "amd64"}.IsWildcard()).To(BeTrue())
	})
	It("joker mimariyi joker platform olarak rapor etmeli", func() {
		Expect(Platform{OS: "linux", Arch: "*"}.IsWildcard()).To(BeTrue())
	})
	It("os/arch olarak serileştirilmeli", func() {
		Expect(Platform{OS: "linux", Arch: "amd64"}.String()).To(Equal("linux/amd64"))
	})

	Specify("bir temel mağaza adı üretebilmeli", func() {
		plat := Platform{OS: "linux", Arch: "amd64"}
		ver := Concrete{Major: 1, Minor: 16, Patch: 3}
		Expect(plat.BaseName(ver)).To(Equal("1.16.3-linux-amd64"))
	})

	Specify("bir arşiv adı üretebilmeli", func() {
		plat := Platform{OS: "linux", Arch: "amd64"}
		ver := Concrete{Major: 1, Minor: 16, Patch: 3}
		Expect(plat.ArchiveName(ver)).To(Equal("envtest-v1.16.3-linux-amd64.tar.gz"))
	})

	Describe("parsing", func() {
		Context("sürüm-platform adları için", func() {
			It("x.y.z-os-arch biçimindeki dizeleri kabul etmeli", func() {
				ver, plat := ExtractWithPlatform(VersionPlatformRE, "1.16.3-linux-amd64")
				Expect(ver).To(Equal(&Concrete{Major: 1, Minor: 16, Patch: 3}))
				Expect(plat).To(Equal(Platform{OS: "linux", Arch: "amd64"}))
			})
			It("anlamsız dizeleri reddetmeli", func() {
				ver, _ := ExtractWithPlatform(VersionPlatformRE, "1.16-linux-amd64")
				Expect(ver).To(BeNil())
			})
		})
		Context("arşiv adları için (controller-tools)", func() {
			It("envtest-vx.y.z-os-arch.tar.gz biçimindeki dizeleri kabul etmeli", func() {
				ver, plat := ExtractWithPlatform(ArchiveRE, "envtest-v1.16.3-linux-amd64.tar.gz")
				Expect(ver).To(Equal(&Concrete{Major: 1, Minor: 16, Patch: 3}))
				Expect(plat).To(Equal(Platform{OS: "linux", Arch: "amd64"}))
			})
			It("anlamsız dizeleri reddetmeli", func() {
				ver, _ := ExtractWithPlatform(ArchiveRE, "envtest-v1.16.3-linux-amd64.tar.sum")
				Expect(ver).To(BeNil())
			})
		})
	})
})

var _ = Describe("Spec yardımcıları", func() {
	Specify("bir spec'i somut bir sürümle doldurabilir", func() {
		spec := Spec{Selector: AnySelector{}} // AnyVersion kullanmayın, böylece değiştirmeyiz
		spec.MakeConcrete(Concrete{Major: 1, Minor: 16})
		Expect(spec.AsConcrete()).To(Equal(&Concrete{Major: 1, Minor: 16}))
	})
	It("en son kontrol için ! ile temel seçici olarak serileştirilmeli", func() {
		spec, err := FromExpr("1.16.*!")
		Expect(err).NotTo(HaveOccurred())
		Expect(spec.String()).To(Equal("1.16.*!"))
	})
	It("en son kontrol değilse temel seçici olarak serileştirilmeli", func() {
		spec, err := FromExpr("1.16.*")
		Expect(err).NotTo(HaveOccurred())
		Expect(spec.String()).To(Equal("1.16.*"))
	})
})
