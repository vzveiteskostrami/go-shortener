package dbf

import (
	"bufio"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/vzveiteskostrami/go-shortener/internal/config"
	"github.com/vzveiteskostrami/go-shortener/internal/logging"
	"github.com/vzveiteskostrami/go-shortener/internal/misc"
)

type FMStorage struct {
	fStore *os.File
	store  map[string]StorageURL
}

func (f *FMStorage) DBFInit() int64 {
	var err error
	if config.Storage.FileName != "" {
		s := filepath.Dir(config.Storage.FileName)
		if s != "" {
			err = os.MkdirAll(s, fs.ModeDir)
			if err != nil {
				logging.S().Panic(err)
			}
		}
		f.fStore, err = os.OpenFile(config.Storage.FileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			logging.S().Panic(err)
		}
		logging.S().Infof("Открыт файл %s для записи и чтения", config.Storage.FileName)
	}
	nextNumFile := f.readStoredData()
	return nextNumFile
}

func (f *FMStorage) readStoredData() int64 {
	var err error
	var nextNum int64 = 0
	storageURLItem := StorageURL{}

	if f.store == nil {
		f.store = make(map[string]StorageURL)
	}

	if misc.IsNil(f.fStore) {
		return 0
	}

	scanner := bufio.NewScanner(f.fStore)

	for scanner.Scan() {
		data := scanner.Bytes()
		err = json.Unmarshal(data, &storageURLItem)
		if err != nil {
			logging.S().Panic(err)
		}
		f.store[storageURLItem.ShortURL] = storageURLItem
		if nextNum <= storageURLItem.UUID {
			nextNum = storageURLItem.UUID + 1
		}
	}
	return nextNum
}

func (f *FMStorage) DBFClose() {
	if f.fStore != nil {
		f.fStore.Close()
	}
}

func (f *FMStorage) DBFSaveLink(storageURLItem *StorageURL) {
	if f.store == nil {
		f.store = make(map[string]StorageURL)
	}

	f.store[storageURLItem.ShortURL] = *storageURLItem
	if f.fStore == nil {
		return
	}

	data, _ := json.Marshal(&storageURLItem)
	// добавляем перенос строки
	data = append(data, '\n')
	_, _ = f.fStore.Write(data)
}

func (f *FMStorage) FindLink(link string, byLink bool) (StorageURL, bool) {
	if byLink {
		url, ok := f.store[link]
		return url, ok
	} else {
		for s, url := range f.store {
			if s == link {
				return url, true
			}
		}
		return StorageURL{}, false
	}
}

func (f *FMStorage) PingDBf(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
