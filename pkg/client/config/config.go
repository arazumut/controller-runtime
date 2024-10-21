/*
2017 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisans uyarınca kullanabilirsiniz.
Lisansın bir kopyasını aşağıdaki adreste bulabilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Geçerli yasa uyarınca veya yazılı olarak kabul edilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamında izin verilen belirli dil kapsamındaki
haklar ve sınırlamalar için Lisansa bakın.
*/

package config

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	logf "sigs.k8s.io/controller-runtime/pkg/internal/log"
)

// KubeconfigFlagName kubeconfig bayrağının adıdır
const KubeconfigFlagName = "kubeconfig"

var (
	kubeconfig string
	log        = logf.RuntimeLog.WithName("client").WithName("config")
)

// init, varsayılan komut satırı FlagSet'e "kubeconfig" bayrağını kaydeder.
// TODO: Bu kaldırılmalıdır, çünkü kullanıcılar için bayrakların yeniden tanımlanmasına yol açabilir,
// eğer kodlarının diğer bölümlerinde komut satırı FlagSet'e "kubeconfig" bayrağını zaten kaydetmişlerse.
func init() {
	RegisterFlags(flag.CommandLine)
}

// RegisterFlags, belirtilen FlagSet'e bayrak değişkenlerini zaten kayıtlı değilse kaydeder.
// Varsayılan komut satırı FlagSet'i kullanır, eğer hiçbiri sağlanmamışsa. Şu anda, yalnızca kubeconfig bayrağını kaydeder.
func RegisterFlags(fs *flag.FlagSet) {
	if fs == nil {
		fs = flag.CommandLine
	}
	if f := fs.Lookup(KubeconfigFlagName); f != nil {
		kubeconfig = f.Value.String()
	} else {
		fs.StringVar(&kubeconfig, KubeconfigFlagName, "", "Bir kubeconfig dosyasının yolları. Sadece küme dışındaysa gereklidir.")
	}
}

// GetConfig, bir Kubernetes API sunucusuyla konuşmak için bir *rest.Config oluşturur.
// Eğer --kubeconfig ayarlanmışsa, o konumdaki kubeconfig dosyasını kullanır. Aksi takdirde,
// küme içinde çalıştığını varsayar ve küme tarafından sağlanan kubeconfig'i kullanır.
//
// Ayrıca Kubernetes denetleyici yöneticisi varsayılanlarına (20 QPS, 30 burst) dayalı olarak daha mantıklı varsayılanlar uygular.
//
// Yapılandırma önceliği:
//
// * Bir dosyaya işaret eden --kubeconfig bayrağı
//
// * Bir dosyaya işaret eden KUBECONFIG ortam değişkeni
//
// * Küme içinde çalışıyorsa küme içi yapılandırma
//
// * $HOME/.kube/config dosyası varsa.
func GetConfig() (*rest.Config, error) {
	return GetConfigWithContext("")
}

// GetConfigWithContext, belirli bir bağlamla bir Kubernetes API sunucusuyla konuşmak için bir *rest.Config oluşturur.
// Eğer --kubeconfig ayarlanmışsa, o konumdaki kubeconfig dosyasını kullanır. Aksi takdirde,
// küme içinde çalıştığını varsayar ve küme tarafından sağlanan kubeconfig'i kullanır.
//
// Ayrıca Kubernetes denetleyici yöneticisi varsayılanlarına (20 QPS, 30 burst) dayalı olarak daha mantıklı varsayılanlar uygular.
//
// Yapılandırma önceliği:
//
// * Bir dosyaya işaret eden --kubeconfig bayrağı
//
// * Bir dosyaya işaret eden KUBECONFIG ortam değişkeni
//
// * Küme içinde çalışıyorsa küme içi yapılandırma
//
// * $HOME/.kube/config dosyası varsa.
func GetConfigWithContext(context string) (*rest.Config, error) {
	cfg, err := loadConfig(context)
	if err != nil {
		return nil, err
	}
	if cfg.QPS == 0.0 {
		cfg.QPS = 20.0
	}
	if cfg.Burst == 0 {
		cfg.Burst = 30
	}
	return cfg, nil
}

// loadInClusterConfig, küme içi Kubernetes istemci yapılandırmasını yüklemek için kullanılan bir işlevdir.
// Bu değişken, yapılandırmanın yüklenme önceliğini test etmeyi mümkün kılar.
var loadInClusterConfig = rest.InClusterConfig

// loadConfig, GetConfig'de belirtilen kurallara göre bir REST Yapılandırması yükler.
func loadConfig(context string) (config *rest.Config, configErr error) {
	// Eğer yapılandırma konumunu belirten bir bayrak belirtilmişse, onu kullan
	if len(kubeconfig) > 0 {
		return loadConfigWithContext("", &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}, context)
	}

	// Eğer önerilen kubeconfig ortam değişkeni belirtilmemişse,
	// küme içi yapılandırmayı dene.
	kubeconfigPath := os.Getenv(clientcmd.RecommendedConfigPathEnvVar)
	if len(kubeconfigPath) == 0 {
		c, err := loadInClusterConfig()
		if err == nil {
			return c, nil
		}

		defer func() {
			if configErr != nil {
				log.Error(err, "küme içi yapılandırma yüklenemedi")
			}
		}()
	}

	// Eğer önerilen kubeconfig ortam değişkeni ayarlanmışsa veya
	// küme içi yapılandırma yoksa, varsayılan önerilen konumları dene.
	//
	// NOT: Varsayılan yapılandırma dosyası konumları için, upstream yalnızca
	// kullanıcının ev dizini için $HOME'u kontrol eder, ancak $HOME ayarlanmamışsa
	// os/user.HomeDir'i de deneyebiliriz.
	//
	// TODO(jlanford): bu upstream'de yapılabilir mi?
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if _, ok := os.LookupEnv("HOME"); !ok {
		u, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("mevcut kullanıcı alınamadı: %w", err)
		}
		loadingRules.Precedence = append(loadingRules.Precedence, filepath.Join(u.HomeDir, clientcmd.RecommendedHomeDir, clientcmd.RecommendedFileName))
	}

	return loadConfigWithContext("", loadingRules, context)
}

func loadConfigWithContext(apiServerURL string, loader clientcmd.ClientConfigLoader, context string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loader,
		&clientcmd.ConfigOverrides{
			ClusterInfo: clientcmdapi.Cluster{
				Server: apiServerURL,
			},
			CurrentContext: context,
		}).ClientConfig()
}

// GetConfigOrDie, bir Kubernetes API sunucusuyla konuşmak için bir *rest.Config oluşturur.
// Eğer --kubeconfig ayarlanmışsa, o konumdaki kubeconfig dosyasını kullanır. Aksi takdirde,
// küme içinde çalıştığını varsayar ve küme tarafından sağlanan kubeconfig'i kullanır.
//
// Eğer rest.Config oluşturulurken bir hata oluşursa, bir hata kaydeder ve çıkar.
func GetConfigOrDie() *rest.Config {
	config, err := GetConfig()
	if err != nil {
		log.Error(err, "kubeconfig alınamadı")
		os.Exit(1)
	}
	return config
}
