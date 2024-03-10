package shorturl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/vzveiteskostrami/go-shortener/internal/auth"
	"github.com/vzveiteskostrami/go-shortener/internal/dbf"
	"github.com/vzveiteskostrami/go-shortener/internal/urlman"
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
	//fmt.Fprintln(os.Stdout, "^^^^^^^^^^^^^^", surls)
	ownerID := r.Context().Value(auth.CPownerID).(int64)

	go func() {
		surl := ""
		for _, data := range surls {
			if url, err := dbf.Store.FindLink(context.Background(), data, true); err == nil {
				if !url.Deleted && url.OWNERID == ownerID {
					surl = data
				}
			}
			if surl != "" {
				urlman.WriteToDel(surl)
				surl = ""
			}
		}
	}()
}

