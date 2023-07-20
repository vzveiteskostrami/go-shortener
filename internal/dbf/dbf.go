package dbf

import (
	"net/http"

	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/go-shortener/internal/config"
)

var Store GSStorage

func MakeStorage() {
	if config.Storage.DBConnect != "" {
		var s PGStorage
		Store = &s
	} else {
		var s FMStorage
		Store = &s
	}
}

type GSStorage interface {
	DBFInit() int64
	DBFClose()
	DBFSaveLink(storageURLItem *StorageURL)
	FindLink(link string, byLink bool) (StorageURL, bool)
	PingDBf(w http.ResponseWriter, r *http.Request)
}

type StorageURL struct {
	UUID        int64  `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
