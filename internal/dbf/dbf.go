package dbf

import (
	"context"
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
	DBFSaveLink(storageURLItem *StorageURL) error
	PingDBf(w http.ResponseWriter, r *http.Request)
	FindLink(ctx context.Context, link string, byLink bool) (StorageURL, error)
	DBFGetOwnURLs(ctx context.Context, ownerID int64) ([]StorageURL, error)

	AddToDel(surl string)
	BeginDel()
	EndDel()
	PrintDBF()
}

type StorageURL struct {
	OWNERID     int64  `json:"ownerid"`
	UUID        int64  `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	Deleted     bool   `json:"deleted"`
}

/*
func TstWg() {
	var wg sync.WaitGroup
	for _, v := range []int{0, 1, 2, 3} {
		wg.Add(1)
		go func() {
			fmt.Print(v)
			wg.Done()
		}()
	}
	wg.Wait()
}
*/
