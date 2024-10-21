package finalizer

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type mockFinalizer struct {
	result Result
	err    error
}

func (f mockFinalizer) Finalize(ctx context.Context, obj client.Object) (Result, error) {
	return f.result, f.err
}

func TestFinalizer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Finalizer Suite")
}

var _ = Describe("TestFinalizer", func() {
	var err error
	var pod *corev1.Pod
	var finalizers Finalizers
	var f mockFinalizer

	BeforeEach(func() {
		pod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{},
		}
		finalizers = NewFinalizers()
		f = mockFinalizer{}
	})

	Describe("Register", func() {
		It("başarıyla bir finalizer kaydeder", func() {
			err = finalizers.Register("finalizers.sigs.k8s.io/testfinalizer", f)
			Expect(err).ToNot(HaveOccurred())
		})

		It("zaten kayıtlı olan bir finalizer kaydetmeye çalışırken hata vermeli", func() {
			err = finalizers.Register("finalizers.sigs.k8s.io/testfinalizer", f)
			Expect(err).ToNot(HaveOccurred())

			// Aynı anahtar ile tekrar Register çağrıldığında hata dönmeli
			err = finalizers.Register("finalizers.sigs.k8s.io/testfinalizer", f)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("zaten kayıtlı"))
		})
	})

	Describe("Finalize", func() {
		It("silme zaman damgası nil ve finalizer yokken başarıyla finalize eder ve Updated true döner", func() {
			err = finalizers.Register("finalizers.sigs.k8s.io/testfinalizer", f)
			Expect(err).ToNot(HaveOccurred())

			pod.DeletionTimestamp = nil
			pod.Finalizers = []string{}

			result, err := finalizers.Finalize(context.TODO(), pod)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.Updated).To(BeTrue())
			// Silme zaman damgası nil ve finalizer yokken, kayıtlı finalizer objeye eklenir
			Expect(pod.Finalizers).To(HaveLen(1))
			Expect(pod.Finalizers[0]).To(Equal("finalizers.sigs.k8s.io/testfinalizer"))
		})

		It("silme zaman damgası var ve finalizer mevcutken başarıyla finalize eder ve Updated true döner", func() {
			now := metav1.Now()
			pod.DeletionTimestamp = &now

			err = finalizers.Register("finalizers.sigs.k8s.io/testfinalizer", f)
			Expect(err).ToNot(HaveOccurred())

			pod.Finalizers = []string{"finalizers.sigs.k8s.io/testfinalizer"}

			result, err := finalizers.Finalize(context.TODO(), pod)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.Updated).To(BeTrue())
			// Başarılı finalize sonrası finalizer objeden kaldırılır
			Expect(pod.Finalizers).To(BeEmpty())
		})

		It("silme zaman damgası nil ve finalizer yokken hata dönmez ve Updated false döner", func() {
			pod.DeletionTimestamp = nil
			pod.Finalizers = []string{}

			result, err := finalizers.Finalize(context.TODO(), pod)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.Updated).To(BeFalse())
			Expect(pod.Finalizers).To(BeEmpty())
		})

		It("silme zaman damgası var ve finalizer yokken hata dönmez ve Updated false döner", func() {
			now := metav1.Now()
			pod.DeletionTimestamp = &now
			pod.Finalizers = []string{}

			result, err := finalizers.Finalize(context.TODO(), pod)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.Updated).To(BeFalse())
			Expect(pod.Finalizers).To(BeEmpty())
		})

		It("silme zaman damgası var ve finalizer mevcutken birden fazla finalizer başarıyla finalize eder ve Updated true döner", func() {
			now := metav1.Now()
			pod.DeletionTimestamp = &now

			err = finalizers.Register("finalizers.sigs.k8s.io/testfinalizer", f)
			Expect(err).ToNot(HaveOccurred())

			err = finalizers.Register("finalizers.sigs.k8s.io/newtestfinalizer", f)
			Expect(err).ToNot(HaveOccurred())

			pod.Finalizers = []string{"finalizers.sigs.k8s.io/testfinalizer", "finalizers.sigs.k8s.io/newtestfinalizer"}

			result, err := finalizers.Finalize(context.TODO(), pod)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.Updated).To(BeTrue())
			Expect(result.StatusUpdated).To(BeFalse())
			Expect(pod.Finalizers).To(BeEmpty())
		})

		It("sonuç false ve hata dönmeli", func() {
			now := metav1.Now()
			pod.DeletionTimestamp = &now
			pod.Finalizers = []string{"finalizers.sigs.k8s.io/testfinalizer"}

			f.result.Updated = false
			f.result.StatusUpdated = false
			f.err = fmt.Errorf("finalizer %q için başarısız oldu", pod.Finalizers[0])

			err = finalizers.Register("finalizers.sigs.k8s.io/testfinalizer", f)
			Expect(err).ToNot(HaveOccurred())

			result, err := finalizers.Finalize(context.TODO(), pod)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("finalizer başarısız oldu"))
			Expect(result.Updated).To(BeFalse())
			Expect(result.StatusUpdated).To(BeFalse())
			Expect(pod.Finalizers).To(HaveLen(1))
			Expect(pod.Finalizers[0]).To(Equal("finalizers.sigs.k8s.io/testfinalizer"))
		})

		It("birden fazla finalizer kaydedildiğinde beklenen sonuç ve hata değerlerini dönmeli", func() {
			now := metav1.Now()
			pod.DeletionTimestamp = &now
			pod.Finalizers = []string{
				"finalizers.sigs.k8s.io/testfinalizer1",
				"finalizers.sigs.k8s.io/testfinalizer2",
				"finalizers.sigs.k8s.io/testfinalizer3",
			}

			// Farklı dönüş değerleri ile birden fazla finalizer kaydediliyor
			// Updated true ve hata nil için test
			f.result.Updated = true
			f.result.StatusUpdated = false
			f.err = nil
			err = finalizers.Register("finalizers.sigs.k8s.io/testfinalizer1", f)
			Expect(err).ToNot(HaveOccurred())

			result, err := finalizers.Finalize(context.TODO(), pod)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.Updated).To(BeTrue())
			Expect(result.StatusUpdated).To(BeFalse())
			// `finalizers.sigs.k8s.io/testfinalizer1` finalizer listesinden kaldırılacak,
			// bu yüzden uzunluk 2 olacak.
			Expect(pod.Finalizers).To(HaveLen(2))
			Expect(pod.Finalizers[0]).To(Equal("finalizers.sigs.k8s.io/testfinalizer2"))
			Expect(pod.Finalizers[1]).To(Equal("finalizers.sigs.k8s.io/testfinalizer3"))

			// Updated ve StatusUpdated false ve hata non-nil için test
			f.result.Updated = false
			f.result.StatusUpdated = false
			f.err = fmt.Errorf("finalizer başarısız oldu")
			err = finalizers.Register("finalizers.sigs.k8s.io/testfinalizer2", f)
			Expect(err).ToNot(HaveOccurred())

			result, err = finalizers.Finalize(context.TODO(), pod)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("finalizer başarısız oldu"))
			Expect(result.Updated).To(BeFalse())
			Expect(result.StatusUpdated).To(BeFalse())
			Expect(pod.Finalizers).To(HaveLen(2))
			Expect(pod.Finalizers[0]).To(Equal("finalizers.sigs.k8s.io/testfinalizer2"))
			Expect(pod.Finalizers[1]).To(Equal("finalizers.sigs.k8s.io/testfinalizer3"))

			// Sonuç true ve hata non-nil için test
			f.result.Updated = true
			f.result.StatusUpdated = true
			f.err = fmt.Errorf("finalizer başarısız oldu")
			err = finalizers.Register("finalizers.sigs.k8s.io/testfinalizer3", f)
			Expect(err).ToNot(HaveOccurred())

			result, err = finalizers.Finalize(context.TODO(), pod)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("finalizer başarısız oldu"))
			Expect(result.Updated).To(BeTrue())
			Expect(result.StatusUpdated).To(BeTrue())
			Expect(pod.Finalizers).To(HaveLen(2))
			Expect(pod.Finalizers[0]).To(Equal("finalizers.sigs.k8s.io/testfinalizer2"))
			Expect(pod.Finalizers[1]).To(Equal("finalizers.sigs.k8s.io/testfinalizer3"))
		})
	})
})
