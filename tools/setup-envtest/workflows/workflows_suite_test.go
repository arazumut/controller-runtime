/*
Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı yalnızca Lisans'a uygun olarak kullanabilirsiniz.
Lisans'ın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izinle aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamındaki izinler ve sınırlamalar hakkında daha fazla bilgi için
Lisans'a bakınız.
*/

package workflows_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var testLog logr.Logger

func zapLogger() logr.Logger {
	testOut := zapcore.AddSync(GinkgoWriter)
	encCfg := zap.NewDevelopmentEncoderConfig()
	enc := zapcore.NewConsoleEncoder(encCfg)
	zapLog := zap.New(zapcore.NewCore(enc, testOut, zap.DebugLevel),
		zap.ErrorOutput(testOut), zap.Development(), zap.AddStacktrace(zap.WarnLevel))
	return zapr.NewLogger(zapLog)
}

func TestWorkflows(t *testing.T) {
	testLog = zapLogger()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Workflows Suite")
}
