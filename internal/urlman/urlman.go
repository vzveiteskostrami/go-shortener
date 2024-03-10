package urlman

import (
	"strconv"
	"sync"

	"github.com/vzveiteskostrami/go-shortener/internal/config"
)

var (
	currURLNum  int64 = 0
	lockCounter sync.Mutex
	lockWrite   sync.Mutex
)

func SetURLNum(num int64) {
	currURLNum = num
}

func MakeURL(num int64) string {
	if config.Addresses.In == nil {
		config.ReadData()
	}
	return config.Addresses.Out.Host + ":" + strconv.Itoa(config.Addresses.Out.Port) + "/" + strconv.FormatInt(num, 36)
}

func HoldNumber() int64 {
	lockCounter.Lock()
	return currURLNum
}

func UnlockNumbers() {
	lockCounter.Unlock()
}

func NumberUsed() int64 {
	currURLNum++
	return currURLNum
}
