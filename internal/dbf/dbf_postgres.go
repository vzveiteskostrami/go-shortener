package dbf

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/go-shortener/internal/logging"
)

type PGStorage struct {
	db         *sql.DB
	delSQLBody string
	// [BLOCKER] если в delSQLParams всегда записывается строка, то можно сделать ее как []string
	// [OBJECTION] У меня тогда перестаёт работать этот execContext. Тип delsQLParams
	// должен быть как минимум []any.
	// _, err := d.db.ExecContext(context.Background(), d.delSQLBody, d.delSQLParams...)
	//
	// cannot use d.delSQLParams (variable of type []string) as []any value in
	// argument to d.db.ExecContextcompiler (IncompatibleAssign)
	//
	delSQLParams []interface{}
}

func (d *PGStorage) DBFGetOwnURLs(ctx context.Context, ownerID int64) ([]StorageURL, error) {
	rows, err := d.db.QueryContext(ctx, "SELECT SHORTURL,ORIGINALURL from urlstore WHERE OWNERID=$1;", ownerID)
	if err != nil {
		logging.S().Error(err)
		return nil, err
	}
	if rows.Err() != nil {
		logging.S().Error(rows.Err())
		return nil, rows.Err()
	}
	defer rows.Close()

	items := make([]StorageURL, 0)
	item := StorageURL{}
	for rows.Next() {
		err = rows.Scan(&item.ShortURL, &item.OriginalURL)
		if err != nil {
			logging.S().Error()
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (d *PGStorage) DBFSaveLink(storageURLItem *StorageURL) error {
	//[LINT] здесь лучше тоже прокидывать context, а не использовать Background
	//[OBJECTION] Здесь происходит сохранение данных, а не получение.
	// Если при получении данных пользователь отвалился, то нам да, незачем продолжать
	// работу. Она никому не нужна. Если он отвалился во время записи данных, нас
	// это мало волнует. У нас есь всё, чтобы сохранить данные, и в отсутствии пользователя.
	su, err := d.FindLink(context.Background(), storageURLItem.OriginalURL, false)
	if err == nil {
		storageURLItem.UUID = su.UUID
		storageURLItem.OWNERID = su.OWNERID
		storageURLItem.ShortURL = su.ShortURL
		storageURLItem.Deleted = su.Deleted
	} else {
		//lockWrite.Lock()
		//defer lockWrite.Unlock()
		_, err := d.db.ExecContext(context.Background(), "INSERT INTO urlstore (OWNERID,UUID,SHORTURL,ORIGINALURL,DELETEFLAG) VALUES ($1,$2,$3,$4,$5);",
			storageURLItem.OWNERID,
			storageURLItem.UUID,
			storageURLItem.ShortURL,
			storageURLItem.OriginalURL,
			storageURLItem.Deleted)
		if err != nil {
			// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
			//fmt.Fprintln(os.Stdout, "Мы здесь!", err.Error())
			logging.S().Error(err)
			return err
		}
		// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
		//else {
		//		fmt.Fprintln(os.Stdout, "Вставка "+storageURLItem.OriginalURL)
		//	}
	}
	return nil
}

func (d *PGStorage) FindLink(ctx context.Context, link string, byLink bool) (StorageURL, error) {
	storageURLItem := StorageURL{}
	sbody := ``
	if byLink {
		sbody = "SELECT OWNERID,UUID,SHORTURL,ORIGINALURL,DELETEFLAG from urlstore WHERE shorturl=$1;"
	} else {
		sbody = "SELECT OWNERID,UUID,SHORTURL,ORIGINALURL,DELETEFLAG from urlstore WHERE originalurl=$1;"
	}
	rows, err := d.db.QueryContext(ctx, sbody, link)
	if err != nil {
		logging.S().Error(err)
		// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
		//fmt.Fprintln(os.Stdout, "оппа!", err)
		return StorageURL{}, err
	}
	if rows.Err() != nil {
		err = rows.Err()
		logging.S().Error(err)
		// сохранён/закомментирован вывод на экран. Необходим для сложных случаев тестирования.
		//fmt.Fprintln(os.Stdout, "Оппа два!", rows.Err().Error())
		return StorageURL{}, err
	}
	defer rows.Close()

	ok := false
	for !ok && rows.Next() {
		err = rows.Scan(&storageURLItem.OWNERID, &storageURLItem.UUID, &storageURLItem.ShortURL, &storageURLItem.OriginalURL, &storageURLItem.Deleted)
		if err != nil {
			logging.S().Panic(err)
		}
		ok = true
	}

	if !ok {
		//Приходится искусственно создавать ошибку, чтобы не возвращать bool
		//Денис считает что bool возвращать плохо. Я спорить не буду.
		err = errors.New("не ошибка, но надо же что-то вернуть")
	}

	return storageURLItem, err
}
