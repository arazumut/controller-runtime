// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 The Kubernetes Authors

package env

import (
	"errors"
	"fmt"
	"os"
)

// Exit, verilen kod ve hata mesajı ile çıkar.
//
// Bu davranışı elde etmek için main fonksiyonunda HandleExitWithCode'u defer ile çağırın.
func Exit(code int, msg string, args ...interface{}) {
	panic(&exitCode{
		code: code,
		err:  fmt.Errorf(msg, args...),
	})
}

// ExitCause, verilen kod ve hata mesajı ile çıkar, ayrıca
// geçirilen temel hatayı da otomatik olarak sarar.
//
// Bu davranışı elde etmek için main fonksiyonunda HandleExitWithCode'u defer ile çağırın.
func ExitCause(code int, err error, msg string, args ...interface{}) {
	args = append(args, err)
	panic(&exitCode{
		code: code,
		err:  fmt.Errorf(msg+": %w", args...),
	})
}

// exitCode, bir panik durumunda verilen kod ve mesaj ile çıkışı belirten bir hatadır.
type exitCode struct {
	code int
	err  error
}

func (c *exitCode) Error() string {
	return fmt.Sprintf("%v (çıkış kodu %d)", c.err, c.code)
}

func (c *exitCode) Unwrap() error {
	return c.err
}

// asExit, verilen (panik) değerinin bir exitCode hatası olup olmadığını kontrol eder
// ve eğer öyleyse verilen işaretçiye depolar. Bu, recover() değerlerinde çalışır.
func asExit(val interface{}, exit **exitCode) bool {
	if val == nil {
		return false
	}
	err, isErr := val.(error)
	if !isErr {
		return false
	}
	if !errors.As(err, exit) {
		return false
	}
	return true
}

// HandleExitWithCode, exitCode türündeki panikleri işler,
// durum mesajını yazdırır ve verilen çıkış kodu ile çıkar,
// veya exitCode hatası değilse yeniden panik yapar.
//
// Bu, main fonksiyonunuzda ilk defer olarak kullanılmalıdır.
func HandleExitWithCode() {
	if cause := recover(); CheckRecover(cause, func(code int, err error) {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(code)
	}) {
		panic(cause)
	}
}

// CheckRecover, cause değerini kontrol eder ve verilen geri çağırmayı
// exitCode hatası ise çağırır. Eğer yeniden panik yapılması gerekiyorsa true döner.
//
// Bu, genellikle testler için kullanılır, normalde HandleExitWithCode kullanırsınız.
func CheckRecover(cause interface{}, cb func(int, error)) bool {
	if cause == nil {
		return false
	}
	var exitErr *exitCode
	if !asExit(cause, &exitErr) {
		// exit hatası değilse yeniden panik yap
		return true
	}

	cb(exitErr.code, exitErr.err)
	return false
}
