// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 The Kubernetes Authors

package env

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/tools/setup-envtest/versions"
)

// orderPlatforms platformları OS ve ardından mimariye göre sıralar.
func orderPlatforms(first, second versions.Platform) bool {
	// OS'ye göre, ardından mimariye göre sırala
	if first.OS != second.OS {
		return first.OS < second.OS
	}
	return first.Arch < second.Arch
}

// PrintFormat fetch ve switch sonuçlarını nasıl yazdıracağını belirtir.
// Bu, bir bayrak olarak doğrudan kullanılabileceği için geçerli bir pflag.Value'dir.
type PrintFormat int

const (
	// PrintOverview insan tarafından okunabilir verileri yazdırır,
	// yol, sürüm, mimari ve varsa checksum dahil.
	PrintOverview PrintFormat = iota
	// PrintPath yalnızca yolu yazdırır, süsleme olmadan.
	PrintPath
	// PrintEnv yolu ilgili ortam değişkeni ile birlikte yazdırır, böylece
	// çıktıyı şu şekilde kaynak olarak kullanabilirsiniz:
	// `source $(fetch-envtest switch -p env 1.20.x)`.
	PrintEnv
)

// String bu değeri bir bayrak olarak yazdırır.
func (f PrintFormat) String() string {
	switch f {
	case PrintOverview:
		return "overview"
	case PrintPath:
		return "path"
	case PrintEnv:
		return "env"
	default:
		panic(fmt.Sprintf("beklenmeyen yazdırma formatı %d", int(f)))
	}
}

// Set bu değerin bir bayrak olarak değerini ayarlar.
func (f *PrintFormat) Set(val string) error {
	switch val {
	case "overview":
		*f = PrintOverview
	case "path":
		*f = PrintPath
	case "env":
		*f = PrintEnv
	default:
		return fmt.Errorf("bilinmeyen yazdırma formatı %q, overview|path|env seçeneklerinden birini kullanın", val)
	}
	return nil
}

// Type bu değerin bir bayrak olarak türüdür.
func (PrintFormat) Type() string {
	return "{overview|path|env}"
}
