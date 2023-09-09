package shorturl

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/vzveiteskostrami/go-shortener/internal/auth"
	"github.com/vzveiteskostrami/go-shortener/internal/dbf"
)

var (
	delCh chan string
)

func GoDel() {
	delCh = make(chan string)
	tickCh := make(chan struct{})
	go func() {
		defer close(tickCh)
		for {
			time.Sleep(300 * time.Millisecond)
			tickCh <- struct{}{}
		}
	}()

	go func() {
		defer close(delCh)
		url := ""
		wasAdd := false
		dbf.Store.BeginDel()
		for {
			select {
			//case <-tick:
			case <-tickCh:
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

func DeleteOwnerURLsListf(w http.ResponseWriter, r *http.Request) {
	// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
	//fmt.Fprintln(os.Stdout, "^^^^^^^^^^^^^^", "DeleteOwnerURLsListf")
	w.Header().Set("Content-Type", "text/plain")

	var surls []string
	if err := json.NewDecoder(r.Body).Decode(&surls); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
	//fmt.Fprintln(os.Stdout, "^^^^^^^^^^^^^^", surls)
	ownerID := r.Context().Value(auth.CPownerID).(int64)

	go func() {
		surl := ""
		for _, data := range surls {
			if url, ok := dbf.Store.FindLink(data, true); ok {
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
