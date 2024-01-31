package dbf

import (
	"context"
	"strconv"

	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/go-shortener/internal/logging"
)

func (d *PGStorage) AddToDel(surl string) {
	if d.delSQLBody != "" {
		d.delSQLBody += ","
	}
	d.delSQLParams = append(d.delSQLParams, surl)
	s := strconv.Itoa(len(d.delSQLParams))
	d.delSQLBody += "($" + s + ",true)"
}

func (d *PGStorage) BeginDel() {
	d.delSQLBody = ""
	d.delSQLParams = make([]interface{}, 0)
}

func (d *PGStorage) EndDel() {
	if d.delSQLBody == "" {
		return
	}
	d.delSQLBody = "update urlstore set deleteflag=tmp.df from (values " +
		d.delSQLBody +
		") as tmp (su,df) where urlstore.shorturl=tmp.su;"
	_, err := d.db.ExecContext(context.Background(), d.delSQLBody, d.delSQLParams...)
	if err != nil {
		logging.S().Error(err)
		logging.S().Error(d.delSQLBody)
		logging.S().Error(d.delSQLParams)
	}
}
