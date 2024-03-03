package trust

import (
	"net/http"

	"github.com/vzveiteskostrami/go-shortener/internal/config"
)

func TrustedHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.Storage.TrustedSubnet == "" {
			http.Error(w, "", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
