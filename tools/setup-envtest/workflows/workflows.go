// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 The Kubernetes Authors

package workflows

import (
	"context"
	"io"

	"github.com/go-logr/logr"

	envp "sigs.k8s.io/controller-runtime/tools/setup-envtest/env"
)

// Use, saklanan sürüm-platform çiftleri hakkında bilgi yazdıran,
// gerekli ve istenirse bunları indiren bir iş akışıdır.
type Use struct {
	UseEnv      bool
	AssetsPath  string
	PrintFormat envp.PrintFormat
}

// Do, bu iş akışını yürütür.
func (f Use) Do(env *envp.Env) {
	ctx := logr.NewContext(context.TODO(), env.Log.WithName("use"))
	env.EnsureBaseDirs(ctx)
	if f.UseEnv {
		// env değişkenini koşulsuz olarak kullan
		if env.PathMatches(f.AssetsPath) {
			env.PrintInfo(f.PrintFormat)
			return
		}
	}
	env.EnsureVersionIsSet(ctx)
	if env.ExistsAndValid() {
		env.PrintInfo(f.PrintFormat)
		return
	}
	if env.NoDownload {
		envp.Exit(2, "bu mimari (%s) için disk üzerinde böyle bir sürüm (%s) yok -- disk üzerinde ne olduğunu görmek için `list -i` komutunu çalıştırmayı deneyin", env.Version, env.Platform)
	}
	env.Fetch(ctx)
	env.PrintInfo(f.PrintFormat)
}

// List, mağazada ve uzaktaki sunucuda verilen filtreyle eşleşen
// sürüm-platform çiftlerini listeleyen bir iş akışıdır.
type List struct{}

// Do, bu iş akışını yürütür.
func (List) Do(env *envp.Env) {
	ctx := logr.NewContext(context.TODO(), env.Log.WithName("list"))
	env.EnsureBaseDirs(ctx)
	env.ListVersions(ctx)
}

// Cleanup, mağazadan verilen filtreyle eşleşen sürüm-platform çiftlerini
// kaldıran bir iş akışıdır.
type Cleanup struct{}

// Do, bu iş akışını yürütür.
func (Cleanup) Do(env *envp.Env) {
	ctx := logr.NewContext(context.TODO(), env.Log.WithName("cleanup"))

	env.NoDownload = true
	env.ForceDownload = false

	env.EnsureBaseDirs(ctx)
	env.Remove(ctx)
}

// Sideload, verilen arşivi dosyalar olarak kullanarak mağazaya bir sürüm-platform
// çifti ekleyen veya değiştiren bir iş akışıdır.
type Sideload struct {
	Input       io.Reader
	PrintFormat envp.PrintFormat
}

// Do, bu iş akışını yürütür.
func (f Sideload) Do(env *envp.Env) {
	ctx := logr.NewContext(context.TODO(), env.Log.WithName("sideload"))

	env.EnsureBaseDirs(ctx)
	env.NoDownload = true
	env.Sideload(ctx, f.Input)
	env.PrintInfo(f.PrintFormat)
}
