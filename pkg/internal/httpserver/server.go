package httpserver

import (
	"net/http"
	"time"
)

// New işlevi, mantıklı varsayılanlarla yeni bir sunucu döndürür.
func New(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:           handler,          // İstekleri işlemek için kullanılan handler
		MaxHeaderBytes:    1 << 20,          // Maksimum başlık boyutu (1 MB)
		IdleTimeout:       90 * time.Second, // Boşta kalma zaman aşımı süresi (90 saniye)
		ReadHeaderTimeout: 32 * time.Second, // Başlık okuma zaman aşımı süresi (32 saniye)
	}
}
