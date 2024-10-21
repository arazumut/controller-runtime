/*
2020 Kubernetes Yazarları tarafından oluşturulmuştur.

Apache Lisansı, Sürüm 2.0 (Lisans) altında lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ OLMAKSIZIN; açık veya zımni garantiler dahil.
Lisans kapsamındaki izin ve sınırlamalar için Lisansa bakınız.
*/

package client

import (
	"context"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
)

// NewWithWatch yeni bir WithWatch döner.
func NewWithWatch(config *rest.Config, options Options) (WithWatch, error) {
	client, err := newClient(config, options)
	if err != nil {
		return nil, err
	}
	return &izlemeClient{client: client}, nil
}

type izlemeClient struct {
	*client
}

func (w *izlemeClient) Watch(ctx context.Context, list ObjectList, opts ...ListOption) (watch.Interface, error) {
	switch l := list.(type) {
	case runtime.Unstructured:
		return w.yapısızIzle(ctx, l, opts...)
	case *metav1.PartialObjectMetadataList:
		return w.metadataIzle(ctx, l, opts...)
	default:
		return w.tipIzle(ctx, l, opts...)
	}
}

func (w *izlemeClient) listeSeçenekleri(opts ...ListOption) ListOptions {
	listOpts := ListOptions{}
	listOpts.ApplyOptions(opts)
	if listOpts.Raw == nil {
		listOpts.Raw = &metav1.ListOptions{}
	}
	listOpts.Raw.Watch = true

	return listOpts
}

func (w *izlemeClient) metadataIzle(ctx context.Context, obj *metav1.PartialObjectMetadataList, opts ...ListOption) (watch.Interface, error) {
	gvk := obj.GroupVersionKind()
	gvk.Kind = strings.TrimSuffix(gvk.Kind, "List")

	listOpts := w.listeSeçenekleri(opts...)

	resInt, err := w.client.metadataClient.getResourceInterface(gvk, listOpts.Namespace)
	if err != nil {
		return nil, err
	}

	return resInt.Watch(ctx, *listOpts.AsListOptions())
}

func (w *izlemeClient) yapısızIzle(ctx context.Context, obj runtime.Unstructured, opts ...ListOption) (watch.Interface, error) {
	r, err := w.client.unstructuredClient.resources.getResource(obj)
	if err != nil {
		return nil, err
	}

	listOpts := w.listeSeçenekleri(opts...)

	return r.Get().
		NamespaceIfScoped(listOpts.Namespace, r.isNamespaced()).
		Resource(r.resource()).
		VersionedParams(listOpts.AsListOptions(), w.client.unstructuredClient.paramCodec).
		Watch(ctx)
}

func (w *izlemeClient) tipIzle(ctx context.Context, obj ObjectList, opts ...ListOption) (watch.Interface, error) {
	r, err := w.client.typedClient.resources.getResource(obj)
	if err != nil {
		return nil, err
	}

	listOpts := w.listeSeçenekleri(opts...)

	return r.Get().
		NamespaceIfScoped(listOpts.Namespace, r.isNamespaced()).
		Resource(r.resource()).
		VersionedParams(listOpts.AsListOptions(), w.client.typedClient.paramCodec).
		Watch(ctx)
}
