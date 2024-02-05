package dbf

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/vzveiteskostrami/go-shortener/internal/auth"
	"github.com/vzveiteskostrami/go-shortener/internal/config"
	"github.com/vzveiteskostrami/go-shortener/internal/logging"
	"github.com/vzveiteskostrami/go-shortener/internal/misc"
)

type FMStorage struct {
	fStore             *os.File
	store              map[string]StorageURL
	needRewriteStorage bool
	makeOp             sync.Mutex
}

func (f *FMStorage) makeStorage() {
	if f.store == nil {
		f.makeOp.Lock()
		defer f.makeOp.Unlock()
		if f.store == nil {
			f.store = make(map[string]StorageURL)
		}
	}
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
	var nextNumFile int64 = 0
	var newOwnerID int64
	nextNumFile, newOwnerID = f.readStoredData()
	auth.SetNewOwnerID(newOwnerID)
	return nextNumFile
}

func (f *FMStorage) readStoredData() (int64, int64) {
	var err error
	var nextNum int64 = 0
	var nextOWNERID int64 = 0
	storageURLItem := StorageURL{}

	f.makeStorage()

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

func (f *FMStorage) DBFSaveLink(storageURLItem *StorageURL) error {
	f.makeStorage()

	data, _ := json.Marshal(&storageURLItem)
	// добавляем перенос строки
	data = append(data, '\n')

	f.makeOp.Lock()
	defer f.makeOp.Unlock()

	f.store[storageURLItem.ShortURL] = *storageURLItem

	if f.fStore == nil {
		return nil
	}

	_, _ = f.fStore.Write(data)

	return nil
}

func (f *FMStorage) BeginDel() {
	f.needRewriteStorage = false
}

func (f *FMStorage) AddToDel(surl string) {
	f.makeStorage()

	st := f.store[surl]
	st.Deleted = true
	f.store[surl] = st
	f.needRewriteStorage = true
}

func (f *FMStorage) EndDel() {
	if f.fStore == nil || !f.needRewriteStorage {
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

func (f *FMStorage) FindLink(ctx context.Context, link string, byLink bool) (StorageURL, error) {
	err := errors.New("Не найдено в списке")
	if byLink {
		url, ok := f.store[link]
		if ok {
			err = nil
		}
		return url, err
	} else {
		for s, url := range f.store {
			if s == link {
				return url, nil
			}
		}
		return StorageURL{}, err
	}
}

func (f *FMStorage) PingDBf(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func (f *FMStorage) PrintDBF() {
}

// [LINT] здесь можно подумать как по другому хранить данные, чтобы поиск был не О(n), а O(1).
// Можно попробовать использовать map[int64][]StorageUrl. Где int64 - ownerID
// [OBJECTION] Да, именно здесь поиск убыстрится. Но везде, где идёт прямой поиск URL он
// замедлится (FindLink, AddToDel). Потому что надо будет перебрать всех овнеров в цикле и
// внутри каждого искать URL
func (f *FMStorage) DBFGetOwnURLs(ctx context.Context, ownerID int64) ([]StorageURL, error) {
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
