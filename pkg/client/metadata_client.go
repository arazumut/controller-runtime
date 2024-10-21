/*
Telif Hakkı 2020 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa tarafından gerekli kılınmadıkça veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN.
Lisans kapsamında izin verilen belirli dil kapsamındaki haklar ve
sınırlamalar için Lisansa bakın.
*/

package client

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/metadata"
)

// metadataClient, API sunucusundan yalnızca meta veri isteklerini okuyan ve yazan bir istemcidir.
type metadataClient struct {
	client     metadata.Interface
	restMapper meta.RESTMapper
}

func (mc *metadataClient) getResourceInterface(gvk schema.GroupVersionKind, ns string) (metadata.ResourceInterface, error) {
	mapping, err := mc.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	if mapping.Scope.Name() == meta.RESTScopeNameRoot {
		return mc.client.Resource(mapping.Resource), nil
	}
	return mc.client.Resource(mapping.Resource).Namespace(ns), nil
}

// Delete, client.Client'i uygular.
func (mc *metadataClient) Delete(ctx context.Context, obj Object, opts ...DeleteOption) error {
	metadata, ok := obj.(*metav1.PartialObjectMetadata)
	if !ok {
		return fmt.Errorf("meta veri istemcisi nesneyi anlamadı: %T", obj)
	}

	resInt, err := mc.getResourceInterface(metadata.GroupVersionKind(), metadata.Namespace)
	if err != nil {
		return err
	}

	deleteOpts := DeleteOptions{}
	deleteOpts.ApplyOptions(opts)

	return resInt.Delete(ctx, metadata.Name, *deleteOpts.AsDeleteOptions())
}

// DeleteAllOf, client.Client'i uygular.
func (mc *metadataClient) DeleteAllOf(ctx context.Context, obj Object, opts ...DeleteAllOfOption) error {
	metadata, ok := obj.(*metav1.PartialObjectMetadata)
	if !ok {
		return fmt.Errorf("meta veri istemcisi nesneyi anlamadı: %T", obj)
	}

	deleteAllOfOpts := DeleteAllOfOptions{}
	deleteAllOfOpts.ApplyOptions(opts)

	resInt, err := mc.getResourceInterface(metadata.GroupVersionKind(), deleteAllOfOpts.ListOptions.Namespace)
	if err != nil {
		return err
	}

	return resInt.DeleteCollection(ctx, *deleteAllOfOpts.AsDeleteOptions(), *deleteAllOfOpts.AsListOptions())
}

// Patch, client.Client'i uygular.
func (mc *metadataClient) Patch(ctx context.Context, obj Object, patch Patch, opts ...PatchOption) error {
	metadata, ok := obj.(*metav1.PartialObjectMetadata)
	if !ok {
		return fmt.Errorf("meta veri istemcisi nesneyi anlamadı: %T", obj)
	}

	gvk := metadata.GroupVersionKind()
	resInt, err := mc.getResourceInterface(gvk, metadata.Namespace)
	if err != nil {
		return err
	}

	data, err := patch.Data(obj)
	if err != nil {
		return err
	}

	patchOpts := &PatchOptions{}
	patchOpts.ApplyOptions(opts)

	res, err := resInt.Patch(ctx, metadata.Name, patch.Type(), data, *patchOpts.AsPatchOptions())
	if err != nil {
		return err
	}
	*metadata = *res
	metadata.SetGroupVersionKind(gvk) // GVK'yi geri yükle, meta verilerde ayarlanmamış
	return nil
}

// Get, client.Client'i uygular.
func (mc *metadataClient) Get(ctx context.Context, key ObjectKey, obj Object, opts ...GetOption) error {
	metadata, ok := obj.(*metav1.PartialObjectMetadata)
	if !ok {
		return fmt.Errorf("meta veri istemcisi nesneyi anlamadı: %T", obj)
	}

	gvk := metadata.GroupVersionKind()

	getOpts := GetOptions{}
	getOpts.ApplyOptions(opts)

	resInt, err := mc.getResourceInterface(gvk, key.Namespace)
	if err != nil {
		return err
	}

	res, err := resInt.Get(ctx, key.Name, *getOpts.AsGetOptions())
	if err != nil {
		return err
	}
	*metadata = *res
	metadata.SetGroupVersionKind(gvk) // GVK'yi geri yükle, meta verilerde ayarlanmamış
	return nil
}

// List, client.Client'i uygular.
func (mc *metadataClient) List(ctx context.Context, obj ObjectList, opts ...ListOption) error {
	metadata, ok := obj.(*metav1.PartialObjectMetadataList)
	if !ok {
		return fmt.Errorf("meta veri istemcisi nesneyi anlamadı: %T", obj)
	}

	gvk := metadata.GroupVersionKind()
	gvk.Kind = strings.TrimSuffix(gvk.Kind, "List")

	listOpts := ListOptions{}
	listOpts.ApplyOptions(opts)

	resInt, err := mc.getResourceInterface(gvk, listOpts.Namespace)
	if err != nil {
		return err
	}

	res, err := resInt.List(ctx, *listOpts.AsListOptions())
	if err != nil {
		return err
	}
	*metadata = *res
	metadata.SetGroupVersionKind(gvk) // GVK'yi geri yükle, meta verilerde ayarlanmamış
	return nil
}

func (mc *metadataClient) PatchSubResource(ctx context.Context, obj Object, subResource string, patch Patch, opts ...SubResourcePatchOption) error {
	metadata, ok := obj.(*metav1.PartialObjectMetadata)
	if !ok {
		return fmt.Errorf("meta veri istemcisi nesneyi anlamadı: %T", obj)
	}

	gvk := metadata.GroupVersionKind()
	resInt, err := mc.getResourceInterface(gvk, metadata.Namespace)
	if err != nil {
		return err
	}

	patchOpts := &SubResourcePatchOptions{}
	patchOpts.ApplyOptions(opts)

	body := obj
	if patchOpts.SubResourceBody != nil {
		body = patchOpts.SubResourceBody
	}

	data, err := patch.Data(body)
	if err != nil {
		return err
	}

	res, err := resInt.Patch(ctx, metadata.Name, patch.Type(), data, *patchOpts.AsPatchOptions(), subResource)
	if err != nil {
		return err
	}

	*metadata = *res
	metadata.SetGroupVersionKind(gvk) // GVK'yi geri yükle, meta verilerde ayarlanmamış
	return nil
}
