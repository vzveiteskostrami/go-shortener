package trust

import (
	"net"
	"net/http"

	"github.com/vzveiteskostrami/go-shortener/internal/config"
)

func TrustedHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.Storage.TrustedIPNet == nil {
			http.Error(w, "", http.StatusForbidden)
			return
		}

		sip := r.Header.Get("X-Real-IP")
		ip := net.ParseIP(sip)
		if ip == nil || !config.Storage.TrustedIPNet.Contains(ip) {
			http.Error(w, "", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
