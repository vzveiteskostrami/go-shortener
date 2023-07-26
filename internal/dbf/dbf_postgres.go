package dbf

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/go-shortener/internal/auth"
	"github.com/vzveiteskostrami/go-shortener/internal/config"
	"github.com/vzveiteskostrami/go-shortener/internal/logging"
)

type PGStorage struct {
	db *sql.DB
}

func (d *PGStorage) DBFInit() int64 {
	var err error

	d.db, err = sql.Open("postgres", config.Storage.DBConnect)
	if err != nil {
		logging.S().Panic(err)
	}
	logging.S().Infof("Объявлено соединение с %s", config.Storage.DBConnect)

	err = d.db.Ping()
	if err != nil {
		logging.S().Panic(err)
	}
	logging.S().Infof("Установлено соединение с %s", config.Storage.DBConnect)
	nextNumDB, err := d.tableInitData()
	if err != nil {
		logging.S().Panic(err)
	}
	return nextNumDB
}

func (d *PGStorage) DBFClose() {
	if d.db != nil {
		d.db.Close()
	}
}

func (d *PGStorage) tableInitData() (int64, error) {
	auth.NextOWNERID = 0
	if d.db == nil {
		return -1, errors.New("база данных не инициализирована")
	}
	_, err := d.db.ExecContext(context.Background(), "CREATE TABLE IF NOT EXISTS urlstore(OWNERID bigint not null,UUID bigint NOT NULL,SHORTURL character varying(1000) NOT NULL,ORIGINALURL character varying(1000) NOT NULL);CREATE UNIQUE INDEX IF NOT EXISTS urlstore1 ON urlstore (ORIGINALURL);")
	if err != nil {
		return -1, err
	}
	logging.S().Infof("Таблица URLSTORE либо существовала, либо создана")
	var (
		mxUUID    sql.NullInt64
		mxOwnerID sql.NullInt64
	)

	row := d.db.QueryRowContext(context.Background(), "SELECT MAX(UUID),MAX(OWNERID) as MX FROM urlstore;")
	if row.Err() != nil {
		return -1, row.Err()
	}

	if err = row.Scan(&mxUUID, &mxOwnerID); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		} else {
			return -1, err
		}
	}

	if mxOwnerID.Valid {
		auth.NextOWNERID = mxOwnerID.Int64 + 1
	}

	if mxUUID.Valid {
		return mxUUID.Int64 + 1, nil
	} else {
		return 0, nil
	}
}

func (d *PGStorage) DBFGetOwnURLs(ownerID int64) ([]StorageURL, error) {
	rows, err := d.db.QueryContext(context.Background(), "SELECT SHORTURL,ORIGINALURL from urlstore WHERE OWNERID=$1;", ownerID)
	if err != nil {
		logging.S().Panic(err)
	}
	if rows.Err() != nil {
		logging.S().Panic(rows.Err())
	}
	defer rows.Close()

	items := make([]StorageURL, 0)
	item := StorageURL{}
	for rows.Next() {
		err = rows.Scan(&item.ShortURL, &item.OriginalURL)
		if err != nil {
			logging.S().Panic(err)
		}
		items = append(items, item)
	}

	return items, nil
}

func (d *PGStorage) DBFSaveLink(storageURLItem *StorageURL) {
	su, ok := d.FindLink(storageURLItem.OriginalURL, false)
	if ok {
		storageURLItem.UUID = su.UUID
		storageURLItem.UUID = su.OWNERID
		storageURLItem.ShortURL = su.ShortURL
	} else {
		_, err := d.db.ExecContext(context.Background(), "INSERT INTO urlstore (OWNERID,UUID,SHORTURL,ORIGINALURL) VALUES ($1,$2,$3,$4);",
			storageURLItem.OWNERID,
			storageURLItem.UUID,
			storageURLItem.ShortURL,
			storageURLItem.OriginalURL)
		if err != nil {
			logging.S().Panic(err)
		}
	}
}

func (d *PGStorage) FindLink(link string, byLink bool) (StorageURL, bool) {
	storageURLItem := StorageURL{}
	sbody := ``
	if byLink {
		sbody = "SELECT OWNERID,UUID,SHORTURL,ORIGINALURL from urlstore WHERE shorturl=$1;"
	} else {
		sbody = "SELECT OWNERID,UUID,SHORTURL,ORIGINALURL from urlstore WHERE originalurl=$1;"
	}
	rows, err := d.db.QueryContext(context.Background(), sbody, link)
	if err != nil {
		logging.S().Panic(err)
	}
	if rows.Err() != nil {
		logging.S().Panic(rows.Err())
	}
	defer rows.Close()

	ok := false
	for !ok && rows.Next() {
		err = rows.Scan(&storageURLItem.OWNERID, &storageURLItem.UUID, &storageURLItem.ShortURL, &storageURLItem.OriginalURL)
		if err != nil {
			logging.S().Panic(err)
		}
		ok = true
	}
	return storageURLItem, ok
}

func (d *PGStorage) PingDBf(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	if d.db == nil {
		http.Error(w, `База данных не открыта`, http.StatusInternalServerError)
		return
	}

	err := d.db.Ping()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
