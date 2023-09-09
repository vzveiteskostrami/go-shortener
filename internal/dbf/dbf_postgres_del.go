package dbf

import (
	"context"
	"strconv"

	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/go-shortener/internal/logging"
)

var delSQLBody string
var delSQLParams []interface{}

func (d *PGStorage) AddToDel(surl string) {
	if delSQLBody != "" {
		delSQLBody += ","
	}
	delSQLParams = append(delSQLParams, surl)
	s := strconv.Itoa(len(delSQLParams))
	delSQLBody += "($" + s + ",true)"
}

func (d *PGStorage) BeginDel() {
	delSQLBody = ""
	delSQLParams = make([]interface{}, 0)
}

func (d *PGStorage) EndDel() {
	if delSQLBody == "" {
		return
	}
	delSQLBody = "update urlstore set deleteflag=tmp.df from (values " +
		delSQLBody +
		") as tmp (su,df) where urlstore.shorturl=tmp.su;"
	_, err := d.db.ExecContext(context.Background(), delSQLBody, delSQLParams...)
	if err != nil {
		logging.S().Error(err)
		logging.S().Error(delSQLBody)
		logging.S().Error(delSQLParams)
	}
}
