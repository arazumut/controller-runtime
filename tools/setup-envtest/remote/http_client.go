// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 The Kubernetes Authors

package remote

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/tools/setup-envtest/versions"
	"sigs.k8s.io/yaml"
)

// DefaultIndexURL, HTTPClient içinde kullanılan varsayılan indekstir.
var DefaultIndexURL = "https://raw.githubusercontent.com/kubernetes-sigs/controller-tools/HEAD/envtest-releases.yaml"

var _ Client = &HTTPClient{}

// HTTPClient, envtest ikili arşivlerinin sürümlerini bir indeks üzerinden HTTP ile almak için kullanılan bir istemcidir.
type HTTPClient struct {
	// Log, loglama yapmamızı sağlar.
	Log logr.Logger

	// IndexURL, indeksin URL'sidir, varsayılan olarak DefaultIndexURL kullanılır.
	IndexURL string
}

// Index, envtest ikili arşivlerinin bir indeksini temsil eder. Örnek:
//
//	releases:
//		v1.28.0:
//			envtest-v1.28.0-darwin-amd64.tar.gz:
//	    		hash: <sha512-hash>
//				selfLink: <envtest ikili dosyalarına sahip arşive bağlantı>
type Index struct {
	// Releases, Kubernetes sürümlerini Release (envtest arşivleri) ile eşleştirir.
	Releases map[string]Release `json:"releases"`
}

// Release, bir arşiv adını bir arşiv ile eşleştirir.
type Release map[string]Archive

// Archive, bir arşive ve onun hash değerine bağlantıyı içerir.
type Archive struct {
	Hash     string `json:"hash"`
	SelfLink string `json:"selfLink"`
}

// ListVersions, indekste mevcut olan tüm araç sürümlerini, desteklenen işletim sistemi/mimari kombinasyonları ve ilgili hash değerleri ile listeler.
//
// Sonuçlar, en yeni sürümler önce olacak şekilde sıralanır.
func (c *HTTPClient) ListVersions(ctx context.Context) ([]versions.Set, error) {
	index, err := c.getIndex(ctx)
	if err != nil {
		return nil, err
	}

	knownVersions := map[versions.Concrete][]versions.PlatformItem{}
	for _, releases := range index.Releases {
		for archiveName, archive := range releases {
			ver, details := versions.ExtractWithPlatform(versions.ArchiveRE, archiveName)
			if ver == nil {
				c.Log.V(1).Info("arşiv atlanıyor -- sürümlü araç arşivi gibi görünmüyor", "name", archiveName)
				continue
			}
			c.Log.V(1).Info("sürüm bulundu", "version", ver, "platform", details)
			knownVersions[*ver] = append(knownVersions[*ver], versions.PlatformItem{
				Platform: details,
				Hash: &versions.Hash{
					Type:     versions.SHA512HashType,
					Encoding: versions.HexHashEncoding,
					Value:    archive.Hash,
				},
			})
		}
	}

	res := make([]versions.Set, 0, len(knownVersions))
	for ver, details := range knownVersions {
		res = append(res, versions.Set{Version: ver, Platforms: details})
	}
	// en yeni olanın ilk sırada olması için ters sırada sıralayın
	sort.Slice(res, func(i, j int) bool {
		first, second := res[i].Version, res[j].Version
		return first.NewerThan(second)
	})

	return res, nil
}

// GetVersion, belirtilen sürümü ve platformu indirir ve çıktıya yazar.
func (c *HTTPClient) GetVersion(ctx context.Context, version versions.Concrete, platform versions.PlatformItem, out io.Writer) error {
	index, err := c.getIndex(ctx)
	if err != nil {
		return err
	}

	var loc *url.URL
	var name string
	for _, releases := range index.Releases {
		for archiveName, archive := range releases {
			ver, details := versions.ExtractWithPlatform(versions.ArchiveRE, archiveName)
			if ver == nil {
				c.Log.V(1).Info("arşiv atlanıyor -- sürümlü araç arşivi gibi görünmüyor", "name", archiveName)
				continue
			}

			if *ver == version && details.OS == platform.OS && details.Arch == platform.Arch {
				loc, err = url.Parse(archive.SelfLink)
				if err != nil {
					return fmt.Errorf("selfLink %q parse edilirken hata oluştu, %w", loc, err)
				}
				name = archiveName
				break
			}
		}
	}
	if name == "" {
		return fmt.Errorf("arşiv bulunamadı %s (%s,%s)", version, platform.OS, platform.Arch)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", loc.String(), nil)
	if err != nil {
		return fmt.Errorf("istek oluşturulamadı %s: %w", name, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("alınamadı %s (%s): %w", name, req.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("alınamadı %s (%s) -- durum %q", name, req.URL, resp.Status)
	}

	return readBody(resp, out, name, platform)
}

// FetchSum, belirtilen sürüm ve platform için checksum değerini alır ve platform öğesine yazar.
func (c *HTTPClient) FetchSum(ctx context.Context, version versions.Concrete, platform *versions.PlatformItem) error {
	index, err := c.getIndex(ctx)
	if err != nil {
		return err
	}

	for _, releases := range index.Releases {
		for archiveName, archive := range releases {
			ver, details := versions.ExtractWithPlatform(versions.ArchiveRE, archiveName)
			if ver == nil {
				c.Log.V(1).Info("arşiv atlanıyor -- sürümlü araç arşivi gibi görünmüyor", "name", archiveName)
				continue
			}

			if *ver == version && details.OS == platform.OS && details.Arch == platform.Arch {
				platform.Hash = &versions.Hash{
					Type:     versions.SHA512HashType,
					Encoding: versions.HexHashEncoding,
					Value:    archive.Hash,
				}
				return nil
			}
		}
	}

	return fmt.Errorf("arşiv bulunamadı %s (%s,%s)", version, platform.OS, platform.Arch)
}

func (c *HTTPClient) getIndex(ctx context.Context) (*Index, error) {
	indexURL := c.IndexURL
	if indexURL == "" {
		indexURL = DefaultIndexURL
	}

	loc, err := url.Parse(indexURL)
	if err != nil {
		return nil, fmt.Errorf("indeks URL'si parse edilemedi: %w", err)
	}

	c.Log.V(1).Info("sürümler listeleniyor", "index", indexURL)

	req, err := http.NewRequestWithContext(ctx, "GET", loc.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("indeks almak için istek oluşturulamadı: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("indeks almak için istek gerçekleştirilemedi: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("indeks alınamadı -- durum %q", resp.Status)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("indeks alınamadı -- gövde okunamadı %w", err)
	}

	var index Index
	if err := yaml.Unmarshal(responseBody, &index); err != nil {
		return nil, fmt.Errorf("indeks parse edilemedi: %w", err)
	}
	return &index, nil
}
