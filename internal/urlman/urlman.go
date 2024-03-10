package urlman

import (
	"strconv"
	"sync"
	"time"

	"github.com/vzveiteskostrami/go-shortener/internal/config"
	"github.com/vzveiteskostrami/go-shortener/internal/dbf"
)

var (
	currURLNum  int64 = 0
	lockCounter sync.Mutex
)

var (
	delCh chan string
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

func WriteToDel(surl string) {
	delCh <- surl
}

func GoDel() {
	delCh = make(chan string)
	tick := time.NewTicker(300 * time.Millisecond)

	go func() {
		defer close(delCh)
		defer tick.Stop()
		url := ""
		wasAdd := false
		dbf.Store.BeginDel()
		for {
			select {
			case <-tick.C:
				if wasAdd {
					dbf.Store.EndDel()
					dbf.Store.BeginDel()
					wasAdd = false
				}
			case url = <-delCh:
				dbf.Store.AddToDel(url)
				wasAdd = true
			}
		}
	}()
}

func DoDel() {
	delCh = make(chan string)
	tick := time.NewTicker(300 * time.Millisecond)

	defer close(delCh)
	defer tick.Stop()
	url := ""
	wasAdd := false
	dbf.Store.BeginDel()
	for {
		select {
		case <-tick.C:
			if wasAdd {
				dbf.Store.EndDel()
				dbf.Store.BeginDel()
				wasAdd = false
			}
		case url = <-delCh:
			dbf.Store.AddToDel(url)
			wasAdd = true
		}
	}
}
