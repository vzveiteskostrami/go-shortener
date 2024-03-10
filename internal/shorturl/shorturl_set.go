package shorturl

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/vzveiteskostrami/go-shortener/internal/auth"
	"github.com/vzveiteskostrami/go-shortener/internal/dbf"
	"github.com/vzveiteskostrami/go-shortener/internal/urlman"
)

func SetLink() http.Handler {
	return http.HandlerFunc(SetLinkf)
}

func SetLinkf(w http.ResponseWriter, r *http.Request) {
	// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
	//fmt.Fprintln(os.Stdout, "**************", "SetLinkf")
	w.Header().Set("Content-Type", "text/plain")
	b, err := io.ReadAll(r.Body)
	if err != nil {
		// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
		//fmt.Fprintln(os.Stdout, "Ошибка", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	url := string(b)
	if url == "" {
		// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
		//fmt.Fprintln(os.Stdout, "Ошибка не указан URL")
		http.Error(w, `Не указан URL`, http.StatusBadRequest)
		return
	}

	// [BLOCKER] зачем здесь lock?
	// [OBJECTION] Для монопольного наращивания переменной currURLNum в случае необходимости.
	// Наращивание происходит именно здесь, а не в базе. В базе только фиксится результат.
	// На возражение, что это надо делать в базе, так как там могут несколько пользователей
	// наращивать счётчик в параллель, есть контрвозражение, что в ТЗ упоминается о монопольной
	// работе сервиса, и поэтому выбрана "надбазная" реализация наращивания счётчика.
	nextNum := urlman.HoldNumber()
	defer urlman.UnlockNumbers()

	ownerID := r.Context().Value(auth.CPownerID)

	su := dbf.StorageURL{OriginalURL: url,
		UUID:     nextNum,
		OWNERID:  ownerID.(int64),
		ShortURL: strconv.FormatInt(nextNum, 36)}
	err = dbf.Store.DBFSaveLink(&su)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		if su.UUID == nextNum {
			w.WriteHeader(http.StatusCreated)
			urlman.NumberUsed()
		} else {
			w.WriteHeader(http.StatusConflict)
		}
		w.Write([]byte(urlman.MakeURL(su.UUID)))
	}
	// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
	//fmt.Fprintln(os.Stdout, "Записан URL", url, "как", su.ShortURL)
}

func SetJSONLink() http.Handler {
	return http.HandlerFunc(SetJSONLinkf)
}

func SetJSONLinkf(w http.ResponseWriter, r *http.Request) {
	// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
	//fmt.Fprintln(os.Stdout, "^^^^^^^^^^^^^^", "SetJSONLinkf")
	w.Header().Set("Content-Type", "application/json")
	var url inURL
	if err := json.NewDecoder(r.Body).Decode(&url); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if url.URL == "" {
		http.Error(w, `Не указан URL`, http.StatusBadRequest)
		return
	}

	var surl outURL
	// [BLOCKER] здесь тоже
	// [OBJECTION] Здесь всё выглядит именно так. Но, во-первых, ниже есть
	// бэтчевое сохранение присланых URL, и там выгоднее запереть всё сразу сверху,
	// чем запирать по одной записи. Поэтому здесь просто сохранено однообразие при использовании
	// DBFSaveLink. А во-вторых, мьютекс здесь вообще бы был лишним (или правильнее, не отпирался бы по defer),
	// если бы счётчик URL наращивался бы здесь гарантировано. Но это один из вариантов.
	// Поэтому требует синхронизации. Или переделки способа получения нового URL.
	nextNum := urlman.HoldNumber()
	defer urlman.UnlockNumbers()

	ownerID := r.Context().Value(auth.CPownerID)

	su := dbf.StorageURL{UUID: nextNum,
		OriginalURL: url.URL,
		OWNERID:     ownerID.(int64),
		ShortURL:    strconv.FormatInt(nextNum, 36)}
	dbf.Store.DBFSaveLink(&su)
	if su.UUID == nextNum {
		w.WriteHeader(http.StatusCreated)
		urlman.NumberUsed()
	} else {
		w.WriteHeader(http.StatusConflict)
	}
	var buf bytes.Buffer
	surl.Result = urlman.MakeURL(su.UUID)

	jsonEncoder := json.NewEncoder(&buf)
	jsonEncoder.Encode(surl)
	w.Write(buf.Bytes())
}

func SetJSONBatchLinkf(w http.ResponseWriter, r *http.Request) {
	// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
	//fmt.Fprintln(os.Stdout, "^^^^^^^^^^^^^^", "SetJSONBatchLinkf")
	w.Header().Set("Content-Type", "application/json")
	var urls []cmnURL
	if err := json.NewDecoder(r.Body).Decode(&urls); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(urls) == 0 {
		http.Error(w, `Не указано никаких данных`, http.StatusBadRequest)
		return
	}

	var surls []cmnURL
	num := urlman.HoldNumber()
	defer urlman.UnlockNumbers()

	ownerID := r.Context().Value(auth.CPownerID)

	for _, url := range urls {
		if *url.OriginalURL != "" {
			shorturl := urlman.MakeURL(num)
			surl := cmnURL{CorrelationID: url.CorrelationID, ShortURL: &shorturl}
			surls = append(surls, surl)
			su := dbf.StorageURL{UUID: num,
				OriginalURL: *url.OriginalURL,
				OWNERID:     ownerID.(int64),
				ShortURL:    strconv.FormatInt(num, 36)}
			dbf.Store.DBFSaveLink(&su)
			if su.UUID == num {
				num = urlman.NumberUsed()
			}
		}
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(surls); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(buf.Bytes())
}
