package dbf

import (
	"bufio"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/vzveiteskostrami/go-shortener/internal/auth"
	"github.com/vzveiteskostrami/go-shortener/internal/config"
	"github.com/vzveiteskostrami/go-shortener/internal/logging"
	"github.com/vzveiteskostrami/go-shortener/internal/misc"
)

type FMStorage struct {
	fStore *os.File
	store  map[string]StorageURL
}

var needRewriteStorage bool

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
	var nextNumFile int64 = 0
	nextNumFile, auth.NextOWNERID = f.readStoredData()
	return nextNumFile
}

func (f *FMStorage) readStoredData() (int64, int64) {
	var err error
	var nextNum int64 = 0
	var nextOWNERID int64 = 0
	storageURLItem := StorageURL{}

	if f.store == nil {
		f.store = make(map[string]StorageURL)
	}

	if misc.IsNil(f.fStore) {
		return 0, 0
	}

	scanner := bufio.NewScanner(f.fStore)

	for scanner.Scan() {
		data := scanner.Bytes()
		err = json.Unmarshal(data, &storageURLItem)
		if err != nil {
			logging.S().Panic(err)
		}
		f.store[storageURLItem.ShortURL] = storageURLItem
		//fmt.Println(storageURLItem.OWNERID, storageURLItem.ShortURL)
		if nextNum <= storageURLItem.UUID {
			nextNum = storageURLItem.UUID + 1
		}
		if nextOWNERID <= storageURLItem.OWNERID {
			nextOWNERID = storageURLItem.OWNERID + 1
		}
	}
	return nextNum, nextOWNERID
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

func (f *FMStorage) BeginDel() {
	needRewriteStorage = false
}

func (f *FMStorage) AddToDel(surl string) {
	if f.store == nil {
		f.store = make(map[string]StorageURL)
	}

	st := f.store[surl]
	st.Deleted = true
	f.store[surl] = st
	needRewriteStorage = true
}

func (f *FMStorage) EndDel() {
	if f.fStore == nil || !needRewriteStorage {
		return
	}
	f.fStore.Close()
	os.Truncate(config.Storage.FileName, 0)
	var err error
	f.fStore, err = os.OpenFile(config.Storage.FileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		logging.S().Panic(err)
	}
	for _, url := range f.store {
		data, _ := json.Marshal(&url)
		// добавляем перенос строки
		data = append(data, '\n')
		_, _ = f.fStore.Write(data)
	}
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

func (f *FMStorage) DBFGetOwnURLs(ownerID int64) ([]StorageURL, error) {
	items := make([]StorageURL, 0)
	item := StorageURL{}
	for _, url := range f.store {
		if url.OWNERID == ownerID {
			item.ShortURL = url.ShortURL
			item.OriginalURL = url.OriginalURL
			item.Deleted = url.Deleted
			items = append(items, item)
		}
	}
	return items, nil
}
