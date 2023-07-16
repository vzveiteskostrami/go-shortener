package dbf

import (
	"bufio"
	"context"
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
	var nextNumFile int64 = 0
	var nextNumDB int64 = 0
	if config.Storage.FileName != "" && config.Storage.DBConnect == "" {
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
		nextNumFile = readStoredData(0)
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
		nextNumDB, err = tableInitData()
		if err != nil {
			logging.S().Panic(err)
		}
		//_ = readStoredData(1)
	}

	if nextNumDB > nextNumFile {
		return nextNumDB
	} else {
		return nextNumFile
	}
}

func DBFClose() {
	if fStore != nil {
		fStore.Close()
	}
	if db != nil {
		db.Close()
	}
}

func readStoredData(mode int8) int64 {
	var err error
	var nextNum int64 = 0
	storageURLItem := StorageURL{}

	if store == nil {
		store = make(map[string]StorageURL)
	}

	if mode == 0 {
		if misc.IsNil(fStore) {
			logging.S().Panic(errors.New("не инициализировано"))
		}

		scanner := bufio.NewScanner(fStore)

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
	} else {
		if misc.IsNil(db) {
			logging.S().Panic(errors.New("не инициализировано"))
		}

		rows, err := db.QueryContext(context.Background(), "SELECT UUID,SHORTURL,ORIGINALURL from urlstore order by uuid;")
		if rows.Err() != nil {
			logging.S().Panic(rows.Err())
		}
		if err != nil {
			logging.S().Panic(err)
		}
		defer rows.Close()

		for rows.Next() {
			err = rows.Scan(&storageURLItem.UUID, &storageURLItem.ShortURL, &storageURLItem.OriginalURL)
			if err != nil {
				logging.S().Panic(err)
			}
			store[storageURLItem.ShortURL] = storageURLItem
			if nextNum <= storageURLItem.UUID {
				nextNum = storageURLItem.UUID + 1
			}
		}
	}
	return nextNum
}

func DBFSaveLink(storageURLItem StorageURL) {
	if fStore != nil {
		if store == nil {
			store = make(map[string]StorageURL)
		}

		store[storageURLItem.ShortURL] = storageURLItem
		data, _ := json.Marshal(&storageURLItem)
		// добавляем перенос строки
		data = append(data, '\n')
		_, _ = fStore.Write(data)
	} else if db != nil {
		_, err := db.ExecContext(context.Background(), "INSERT INTO urlstore (UUID,SHORTURL,ORIGINALURL) VALUES ($1,$2,$3);",
			storageURLItem.UUID,
			storageURLItem.ShortURL,
			storageURLItem.OriginalURL)
		if err != nil {
			logging.S().Panic(err)
		}
	}
}

func FindLink(link string) (StorageURL, bool) {
	if db != nil {
		storageURLItem := StorageURL{}
		rows, err := db.QueryContext(context.Background(), "SELECT UUID,SHORTURL,ORIGINALURL from urlstore WHERE uuid=$1;", link)
		if rows.Err() != nil {
			logging.S().Panic(rows.Err())
		}
		if err != nil {
			logging.S().Panic(err)
		}
		defer rows.Close()

		ok := false
		for !ok && rows.Next() {
			err = rows.Scan(&storageURLItem.UUID, &storageURLItem.ShortURL, &storageURLItem.OriginalURL)
			if err != nil {
				logging.S().Panic(err)
			}
			ok = true
		}
		return storageURLItem, ok
	} else {
		url, ok := store[link]
		return url, ok
	}
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

func tableInitData() (int64, error) {
	if db == nil {
		return -1, errors.New("база данных не инициализирована")
	}
	_, err := db.ExecContext(context.Background(), "CREATE TABLE IF NOT EXISTS urlstore(UUID bigint NOT NULL,SHORTURL character varying(1000) NOT NULL,ORIGINALURL character varying(1000) NOT NULL);")
	if err != nil {
		return -1, err
	}
	logging.S().Infof("Таблица URLSTORE либо существовала, либо создана")
	var mx sql.NullInt64

	row := db.QueryRowContext(context.Background(), "SELECT MAX(UUID) as MX FROM urlstore;")
	if row.Err() != nil {
		return -1, row.Err()
	}

	if err = row.Scan(&mx); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		} else {
			return -1, err
		}
	}

	if mx.Valid {
		return mx.Int64 + 1, nil
	} else {
		return 0, nil
	}
}
