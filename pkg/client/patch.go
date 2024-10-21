/*
2018 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisans'a uygun olarak kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa uyarınca veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ OLMAKSIZIN, açık veya zımni olarak.
Lisans kapsamında izin verilen özel dildeki haklar ve
sınırlamalar için Lisansa bakınız.
*/

package client

import (
	"fmt"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

var (
	// Apply verilen nesneyi sunucu tarafında uygulamak için kullanılır.
	Apply Patch = applyPatch{}

	// Merge ham nesneyi herhangi bir değişiklik yapmadan birleştirme yaması olarak kullanır.
	// Fark hesaplamak istiyorsanız MergeFrom kullanın.
	Merge Patch = mergePatch{}
)

type patch struct {
	patchType types.PatchType
	data      []byte
}

// Type Patch'i uygular.
func (s *patch) Type() types.PatchType {
	return s.patchType
}

// Data Patch'i uygular.
func (s *patch) Data(obj Object) ([]byte, error) {
	return s.data, nil
}

// RawPatch verilen PatchType ve veri ile yeni bir Yama oluşturur.
func RawPatch(patchType types.PatchType, data []byte) Patch {
	return &patch{patchType, data}
}

// MergeFromWithOptimisticLock, istemcilerin bir yamanın
// bir nesnenin en son kaynak sürümüne uygulandığından emin olmak istediklerinde kullanılabilir.
//
// Davranış, tüm nesneyi göndermeye gerek kalmadan bir Güncellemenin yapacağına benzer.
// Genellikle bu yöntem, aynı nesne ve aynı API sürümü üzerinde
// ancak farklı Go yapı sürümleriyle hareket eden birden fazla istemciniz olabileceğinde kullanışlıdır.
//
// Örneğin, A ve B alanlarına sahip bir Widget'ın "eski" bir kopyası ve A, B ve C'ye sahip "yeni" bir kopyası.
// Eski yapı tanımını kullanarak bir güncelleme göndermek C'nin düşmesine neden olurken, bir yama kullanmak bunu yapmaz.
type MergeFromWithOptimisticLock struct{}

// ApplyToMergeFrom bu yapılandırmayı verilen yama seçeneklerine uygular.
func (m MergeFromWithOptimisticLock) ApplyToMergeFrom(in *MergeFromOptions) {
	in.OptimisticLock = true
}

// MergeFromOption, birleştirme-yama verileri için seçenekleri değiştiren bazı yapılandırmalardır.
type MergeFromOption interface {
	// ApplyToMergeFrom bu yapılandırmayı verilen yama seçeneklerine uygular.
	ApplyToMergeFrom(*MergeFromOptions)
}

// MergeFromOptions, birleştirme-yama verileri oluşturmak için seçenekler içerir.
type MergeFromOptions struct {
	// OptimisticLock, true olduğunda `metadata.resourceVersion`'ı son yama verilerine dahil eder.
	// `resourceVersion` alanı saklananla eşleşmezse, işlem bir çakışma ile sonuçlanır ve istemcilerin tekrar denemesi gerekir.
	OptimisticLock bool
}

type mergeFromPatch struct {
	patchType   types.PatchType
	createPatch func(originalJSON, modifiedJSON []byte, dataStruct interface{}) ([]byte, error)
	from        Object
	opts        MergeFromOptions
}

// Type Patch'i uygular.
func (s *mergeFromPatch) Type() types.PatchType {
	return s.patchType
}

// Data Patch'i uygular.
func (s *mergeFromPatch) Data(obj Object) ([]byte, error) {
	original := s.from
	modified := obj

	if s.opts.OptimisticLock {
		version := original.GetResourceVersion()
		if len(version) == 0 {
			return nil, fmt.Errorf("OptimisticLock kullanılamaz, nesne %q kullanabileceğimiz herhangi bir kaynak sürümüne sahip değil", original)
		}

		original = original.DeepCopyObject().(Object)
		original.SetResourceVersion("")

		modified = modified.DeepCopyObject().(Object)
		modified.SetResourceVersion(version)
	}

	originalJSON, err := json.Marshal(original)
	if err != nil {
		return nil, err
	}

	modifiedJSON, err := json.Marshal(modified)
	if err != nil {
		return nil, err
	}

	data, err := s.createPatch(originalJSON, modifiedJSON, obj)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func createMergePatch(originalJSON, modifiedJSON []byte, _ interface{}) ([]byte, error) {
	return jsonpatch.CreateMergePatch(originalJSON, modifiedJSON)
}

func createStrategicMergePatch(originalJSON, modifiedJSON []byte, dataStruct interface{}) ([]byte, error) {
	return strategicpatch.CreateTwoWayMergePatch(originalJSON, modifiedJSON, dataStruct)
}

// MergeFrom verilen nesneyi temel alarak birleştirme-yama stratejisi kullanarak bir Yama oluşturur.
// MergeFrom ve StrategicMergeFrom arasındaki fark, değiştirilmiş liste alanlarının işlenmesinde yatar.
// MergeFrom kullanıldığında, mevcut listeler yeni listelerle tamamen değiştirilir.
// StrategicMergeFrom kullanıldığında, liste alanının `patchStrategy`'si API türünde belirtilmişse dikkate alınır,
// örneğin mevcut liste tamamen değiştirilmez, bunun yerine listenin `patchMergeKey`'i kullanılarak yeni liste ile birleştirilir.
// Daha fazla ayrıntı için https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/ adresine bakın.
func MergeFrom(obj Object) Patch {
	return &mergeFromPatch{patchType: types.MergePatchType, createPatch: createMergePatch, from: obj}
}

// MergeFromWithOptions verilen nesneyi temel alarak birleştirme-yama stratejisi kullanarak bir Yama oluşturur.
// Daha fazla ayrıntı için MergeFrom'a bakın.
func MergeFromWithOptions(obj Object, opts ...MergeFromOption) Patch {
	options := &MergeFromOptions{}
	for _, opt := range opts {
		opt.ApplyToMergeFrom(options)
	}
	return &mergeFromPatch{patchType: types.MergePatchType, createPatch: createMergePatch, from: obj, opts: *options}
}

// StrategicMergeFrom verilen nesneyi temel alarak stratejik-birleştirme-yama stratejisi kullanarak bir Yama oluşturur.
// MergeFrom ve StrategicMergeFrom arasındaki fark, değiştirilmiş liste alanlarının işlenmesinde yatar.
// MergeFrom kullanıldığında, mevcut listeler yeni listelerle tamamen değiştirilir.
// StrategicMergeFrom kullanıldığında, liste alanının `patchStrategy`'si API türünde belirtilmişse dikkate alınır,
// örneğin mevcut liste tamamen değiştirilmez, bunun yerine listenin `patchMergeKey`'i kullanılarak yeni liste ile birleştirilir.
// Daha fazla ayrıntı için https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/ adresine bakın.
// Lütfen CRD'lerin stratejik-birleştirme-yama desteklemediğini unutmayın, bkz.
// https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#advanced-features-and-flexibility
func StrategicMergeFrom(obj Object, opts ...MergeFromOption) Patch {
	options := &MergeFromOptions{}
	for _, opt := range opts {
		opt.ApplyToMergeFrom(options)
	}
	return &mergeFromPatch{patchType: types.StrategicMergePatchType, createPatch: createStrategicMergePatch, from: obj, opts: *options}
}

// mergePatch nesneyi yama stratejisi kullanarak birleştirir.
type mergePatch struct{}

// Type Patch'i uygular.
func (p mergePatch) Type() types.PatchType {
	return types.MergePatchType
}

// Data Patch'i uygular.
func (p mergePatch) Data(obj Object) ([]byte, error) {
	// NB(directxman12): burada teknik olarak gerçek bir kodlayıcı kullanmak isteyebiliriz
	// (daha performanslı bir kodlayıcı tanıtılırsa) ancak bu
	// doğru ve bizim kullanımımız için yeterlidir (client-go'daki JSON kodlayıcı da bunu yapar).
	return json.Marshal(obj)
}

// applyPatch nesneyi sunucu tarafında uygulamak için kullanılır.
type applyPatch struct{}

// Type Patch'i uygular.
func (p applyPatch) Type() types.PatchType {
	return types.ApplyPatchType
}

// Data Patch'i uygular.
func (p applyPatch) Data(obj Object) ([]byte, error) {
	// NB(directxman12): burada teknik olarak gerçek bir kodlayıcı kullanmak isteyebiliriz
	// (daha performanslı bir kodlayıcı tanıtılırsa) ancak bu
	// doğru ve bizim kullanımımız için yeterlidir (client-go'daki JSON kodlayıcı da bunu yapar).
	return json.Marshal(obj)
}
