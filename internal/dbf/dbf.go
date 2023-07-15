package dbf

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/go-shortener/internal/config"
	"github.com/vzveiteskostrami/go-shortener/internal/logging"
	"github.com/vzveiteskostrami/go-shortener/internal/misc"
)

type StorageURL struct {
	UUID        int64  `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

var (
	fStore *os.File
	store  map[string]StorageURL
	db     *sql.DB
)

func DBFInit() int64 {
	var nextNum int64 = 0
	if config.Storage.FileName != "" {
		var err error
		s := filepath.Dir(config.Storage.FileName)
		if s != "" {
			err = os.MkdirAll(s, fs.ModeDir)
			if err != nil {
				logging.S().Panic(err)
			}
		}
		fStore, err = os.OpenFile(config.Storage.FileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			logging.S().Panic(err)
		}
		nextNum = readStoredData()
		logging.S().Infof("Открыт файл %s для записи и чтения", config.Storage.FileName)
	}

	if config.Storage.DBConnect != "" {
		var err error

		db, err = sql.Open("postgres", config.Storage.DBConnect)
		if err != nil {
			logging.S().Panic(err)
		}
		logging.S().Infof("Объявлено соединение с %s", config.Storage.DBConnect)

		err = db.Ping()
		if err != nil {
			logging.S().Panic(err)
		}
		logging.S().Infof("Установлено соединение с %s", config.Storage.DBConnect)
	}

	return nextNum
}

func DBFClose() {
	if fStore != nil {
		fStore.Close()
	}
	if db != nil {
		db.Close()
	}
}

func readStoredData() int64 {

	if misc.IsNil(fStore) {
		logging.S().Panic(errors.New("не инициализировано"))
	}

	scanner := bufio.NewScanner(fStore)
	storageURLItem := StorageURL{}
	var err error
	var nextNum int64 = 0

	if store == nil {
		store = make(map[string]StorageURL)
	}

	for scanner.Scan() {
		data := scanner.Bytes()
		err = json.Unmarshal(data, &storageURLItem)
		if err != nil {
			logging.S().Panic(err)
		}
		store[storageURLItem.ShortURL] = storageURLItem
		if nextNum <= storageURLItem.UUID {
			nextNum = storageURLItem.UUID + 1
		}
	}
	return nextNum
}

func DBFSaveLink(storageURLItem StorageURL) {
	if store == nil {
		store = make(map[string]StorageURL)
	}
	store[storageURLItem.ShortURL] = storageURLItem

	if fStore == nil {
		return
	}

	data, _ := json.Marshal(&storageURLItem)
	// добавляем перенос строки
	data = append(data, '\n')
	_, _ = fStore.Write(data)
}

func FindLink(link string) (StorageURL, bool) {
	url, ok := store[link]
	return url, ok
}

func PingDBf(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	if db == nil {
		http.Error(w, `База данных не открыта`, http.StatusInternalServerError)
		return
	}

	err := db.Ping()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
