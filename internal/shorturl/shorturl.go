package shorturl

import (
	"net/http"

	"github.com/vzveiteskostrami/go-shortener/internal/dbf"
)

type inURL struct {
	URL string `json:"url"`
}

type outURL struct {
	Result string `json:"result"`
}

type cmnURL struct {
	CorrelationID *string `json:"correlation_id,omitempty"`
	OriginalURL   *string `json:"original_url,omitempty"`
	ShortURL      *string `json:"short_url,omitempty"`
	Deleted       *bool   `json:"deleted,omitempty"`
}

func PingDBff(w http.ResponseWriter, r *http.Request) {
	code, err := dbf.Store.PingDBf()

	w.Header().Set("Content-Type", "text/plain")
	if err != nil {
		http.Error(w, err.Error(), code)
	}
	//w.WriteHeader(http.StatusOK)
	return

}
