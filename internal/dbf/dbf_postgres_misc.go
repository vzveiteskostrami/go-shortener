package dbf

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/go-shortener/internal/auth"
	"github.com/vzveiteskostrami/go-shortener/internal/config"
	"github.com/vzveiteskostrami/go-shortener/internal/logging"
)

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
	auth.SetNewOwnerID(0)
	if d.db == nil {
		return -1, errors.New("база данных не инициализирована")
	}
	_, err := d.db.ExecContext(context.Background(), "CREATE TABLE IF NOT EXISTS urlstore(OWNERID bigint not null,UUID bigint NOT NULL,SHORTURL character varying(1000) NOT NULL,ORIGINALURL character varying(1000) NOT NULL,DELETEFLAG boolean DEFAULT false);CREATE UNIQUE INDEX IF NOT EXISTS urlstore1 ON urlstore (ORIGINALURL);")
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
		auth.SetNewOwnerID(mxOwnerID.Int64 + 1)
	}

	if mxUUID.Valid {
		return mxUUID.Int64 + 1, nil
	} else {
		return 0, nil
	}
}

/*
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
*/

func (d *PGStorage) PingDBf() (int, error) {
	if d.db == nil {
		return http.StatusInternalServerError, errors.New(`база данных не открыта`)
	}

	err := d.db.Ping()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func (d *PGStorage) PrintDBF() {
	rows, err := d.db.QueryContext(context.Background(), "SELECT OWNERID,SHORTURL,ORIGINALURL from urlstore;")
	if err != nil {
		logging.S().Panic(err)
	}
	if rows.Err() != nil {
		logging.S().Panic(rows.Err())
	}
	defer rows.Close()

	var ow int64
	var sho string
	var fu string
	logging.S().Infow("--------------")
	for rows.Next() {
		err = rows.Scan(&ow, &sho, &fu)
		if err != nil {
			logging.S().Panic(err)
		}
		logging.S().Infow("", "owher", strconv.FormatInt(ow, 10), "short", sho, "full", fu)
	}
	logging.S().Infow("`````````````")
}
