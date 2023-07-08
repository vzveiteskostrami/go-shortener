package dbf

import (
	"bufio"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

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
)

func DBFInit() int64 {
	var nextNum int64 = 0
	if config.FileStorage.FileName != "" {
		var err error
		s := filepath.Dir(config.FileStorage.FileName)
		if s != "" {
			err = os.MkdirAll(s, fs.ModeDir)
			if err != nil {
				logging.S().Panic(err)
			}
		}
		fStore, err = os.OpenFile(config.FileStorage.FileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			logging.S().Panic(err)
		}
		nextNum = readStoredData()
	}
	return nextNum
}

func DBFClose() {
	fStore.Close()
}

func readStoredData() int64 {

	if misc.IsNil(fStore) {
		logging.S().Panic(errors.New("не инициализировано"))
	}

	scanner := bufio.NewScanner(fStore)
	sho := StorageURL{}
	var err error
	var nextNum int64 = 0

	if store == nil {
		store = make(map[string]StorageURL)
	}

	for scanner.Scan() {
		//fmt.Println(scanner.Text())
		data := scanner.Bytes()
		err = json.Unmarshal(data, &sho)
		if err != nil {
			logging.S().Panic(err)
		}
		store[sho.ShortURL] = sho
		if nextNum <= sho.UUID {
			nextNum = sho.UUID + 1
		}
	}
	return nextNum
}

func DBFSaveLink(sho StorageURL) {
	if store == nil {
		store = make(map[string]StorageURL)
	}
	store[sho.ShortURL] = sho

	if fStore == nil {
		return
	}

	data, _ := json.Marshal(&sho)
	// добавляем перенос строки
	data = append(data, '\n')
	_, _ = fStore.Write(data)
}

func FindLink(link string) (StorageURL, bool) {
	url, ok := store[link]
	return url, ok
}
