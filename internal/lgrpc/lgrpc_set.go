package lgrpc

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/vzveiteskostrami/go-shortener/internal/dbf"
	pb "github.com/vzveiteskostrami/go-shortener/internal/proto"
	"github.com/vzveiteskostrami/go-shortener/internal/urlman"
)

func (s *GRPCServer) SetBatchUrl(c context.Context, b *pb.SetBatchUrlRequest) (r *pb.SetBatchUrlResponse, err error) {
	ownerID, token := getReqParams(c)

	r = &pb.SetBatchUrlResponse{}
	r.R = &pb.Resp{Code: http.StatusOK, Msg: "", Token: token}

	num := urlman.HoldNumber()
	defer urlman.UnlockNumbers()

	for _, url := range b.Urls {
		if url.Url != "" {
			su := dbf.StorageURL{UUID: num,
				OriginalURL: url.Url,
				OWNERID:     ownerID,
				ShortURL:    strconv.FormatInt(num, 36)}
			fmt.Println("save=", su)
			dbf.Store.DBFSaveLink(&su)
			if su.UUID == num {
				num = urlman.NumberUsed()
			}
			shorturl := urlman.MakeURL(su.UUID)
			surl := &pb.Curl{CorrelationId: url.CorrelationId, Url: shorturl}
			r.Urls = append(r.Urls, surl)
		}
	}
	return
}

func (s *GRPCServer) SetLink(c context.Context, l *pb.SetLinkRequest) (r *pb.SetLinkResponse, err error) {
	ownerID, token := getReqParams(c)

	r = &pb.SetLinkResponse{}
	r.R = &pb.Resp{Code: http.StatusOK, Msg: "", Token: token}

	if l.Url == "" {
		r.R.Code = http.StatusBadRequest
		r.R.Msg = "Не указан URL"
		return
	}

	nextNum := urlman.HoldNumber()
	defer urlman.UnlockNumbers()

	su := dbf.StorageURL{OriginalURL: l.Url,
		UUID:     nextNum,
		OWNERID:  ownerID,
		ShortURL: strconv.FormatInt(nextNum, 36)}
	err = dbf.Store.DBFSaveLink(&su)
	if err != nil {
		r.R.Code = http.StatusBadRequest
		r.R.Msg = err.Error()
		return
	} else {
		if su.UUID == nextNum {
			r.R.Code = http.StatusCreated
			urlman.NumberUsed()
		} else {
			r.R.Code = http.StatusConflict
			r.R.Msg = "url уже сохранён"
		}
		r.Shorturl = urlman.MakeURL(su.UUID)
	}
	return
}
