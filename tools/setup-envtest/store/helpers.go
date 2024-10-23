// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 The Kubernetes Authors

package store

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// VarsayılanDepoDizini, depo için varsayılan konumu döndürür.
// İşletim sistemine bağlıdır:
//
// - Windows: %LocalAppData%\kubebuilder-envtest
// - OSX: ~/Library/Application Support/io.kubebuilder.envtest
// - Diğerleri: ${XDG_DATA_HOME:-~/.local/share}/kubebuilder-envtest
//
// Aksi takdirde, hata döner. Bu yolların manuel olarak güvenilmemesi gerektiğini unutmayın.
func VarsayılanDepoDizini() (string, error) {
	var temelDizin string

	// temel veri dizinini bul
	switch runtime.GOOS {
	case "windows":
		temelDizin = os.Getenv("LocalAppData")
		if temelDizin == "" {
			return "", errors.New("%LocalAppData% tanımlı değil")
		}
	case "darwin", "ios":
		evDizini := os.Getenv("HOME")
		if evDizini == "" {
			return "", errors.New("$HOME tanımlı değil")
		}
		temelDizin = filepath.Join(evDizini, "Library/Application Support")
	default:
		temelDizin = os.Getenv("XDG_DATA_HOME")
		if temelDizin == "" {
			evDizini := os.Getenv("HOME")
			if evDizini == "" {
				return "", errors.New("ne $XDG_DATA_HOME ne de $HOME tanımlı değil")
			}
			temelDizin = filepath.Join(evDizini, ".local/share")
		}
	}

	// programımıza özgü dizini ekle (OSX biraz farklı bir konvansiyona sahip, bu yüzden onu takip etmeye çalışın).
	switch runtime.GOOS {
	case "darwin", "ios":
		return filepath.Join(temelDizin, "io.kubebuilder.envtest"), nil
	default:
		return filepath.Join(temelDizin, "kubebuilder-envtest"), nil
	}
}
