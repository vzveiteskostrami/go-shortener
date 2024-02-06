package shorturl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/vzveiteskostrami/go-shortener/internal/auth"
	"github.com/vzveiteskostrami/go-shortener/internal/dbf"
)

var (
	delCh chan string
)

func DeleteOwnerURLsListf(w http.ResponseWriter, r *http.Request) {
	// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
	fmt.Fprintln(os.Stdout, "^^^^^^^^^^^^^^", "DeleteOwnerURLsListf")
	w.Header().Set("Content-Type", "text/plain")

	var surls []string
	if err := json.NewDecoder(r.Body).Decode(&surls); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
	fmt.Fprintln(os.Stdout, "^^^^^^^^^^^^^^", surls)
	ownerID := r.Context().Value(auth.CPownerID).(int64)

	go func() {
		surl := ""
		for _, data := range surls {
			if url, ok := dbf.Store.FindLink(context.Background(), data, true); ok {
				if !url.Deleted && url.OWNERID == ownerID {
					surl = data
				}
			}
			if surl != "" {
				delCh <- surl
				surl = ""
			}
		}
	}()
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
