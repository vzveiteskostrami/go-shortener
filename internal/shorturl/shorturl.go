package shorturl

import (
	"strconv"
	"sync"

	"github.com/vzveiteskostrami/go-shortener/internal/config"
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

var (
	currURLNum  int64 = 0
	lockCounter sync.Mutex
	lockWrite   sync.Mutex
)

func SetURLNum(num int64) {
	currURLNum = num
}

func makeURL(num int64) string {
	if config.Addresses.In == nil {
		config.ReadData()
	}
	return config.Addresses.Out.Host + ":" + strconv.Itoa(config.Addresses.Out.Port) + "/" + strconv.FormatInt(num, 36)
}
