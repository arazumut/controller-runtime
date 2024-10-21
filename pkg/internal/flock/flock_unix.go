//go:build linux || darwin || freebsd || openbsd || netbsd || dragonfly
// +build linux darwin freebsd openbsd netbsd dragonfly

/*
Telif Hakkı 2016 Kubernetes Yazarları.

Apache Lisansı, Sürüm 2.0 ("Lisans") uyarınca lisanslanmıştır;
bu dosyayı ancak Lisans'a uygun şekilde kullanabilirsiniz.
Lisans'ın bir kopyasını aşağıdaki adresten edinebilirsiniz:

	http://www.apache.org/licenses/LICENSE-2.0

Yürürlükteki yasa veya yazılı izin gereği aksi belirtilmedikçe,
Lisans kapsamında dağıtılan yazılım "OLDUĞU GİBİ" dağıtılır,
HERHANGİ BİR GARANTİ VEYA KOŞUL OLMAKSIZIN, açık veya zımni.
Lisans kapsamındaki izin ve sınırlamalar hakkında daha fazla bilgi için
Lisans'a bakınız.
*/

package flock

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// Acquire, bir dosya üzerinde işlem süresince kilit alır. Bu yöntem
// yeniden girişimlidir.
func Acquire(path string) error {
	fd, err := unix.Open(path, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC, 0600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("dosya %q kilitlenemiyor: %w", path, ErrAlreadyLocked)
		}
		return err
	}

	// Dosya tanıtıcısını kapatmamıza gerek yok çünkü
	// işlem sona erene kadar onu tutmalıyız.
	err = unix.Flock(fd, unix.LOCK_NB|unix.LOCK_EX)
	if errors.Is(err, unix.EWOULDBLOCK) { // Bu koşul LOCK_NB gerektirir.
		return fmt.Errorf("dosya %q kilitlenemiyor: %w", path, ErrAlreadyLocked)
	}
	return err
}
