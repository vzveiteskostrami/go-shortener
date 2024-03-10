package lgrpc

import (
	"context"
	"net/http"

	"github.com/vzveiteskostrami/go-shortener/internal/dbf"
	pb "github.com/vzveiteskostrami/go-shortener/internal/proto"
)

func (s *GRPCServer) GetLink(c context.Context, g *pb.GetLinkRequest) (r *pb.GetLinkResponse, err error) {
	r = &pb.GetLinkResponse{}
	r.R = &pb.Resp{Code: http.StatusOK, Msg: ""}

	sl := getFuncName(g.Surl)

	url, e := dbf.Store.FindLink(c, sl, true)
	if e != nil {
		r.R.Code = http.StatusBadRequest
		r.R.Msg = "Не найден shortURL"
	} else {
		if url.Deleted {
			r.R.Code = http.StatusGone
		} else {
			r.Url = url.OriginalURL
			r.R.Code = http.StatusTemporaryRedirect
		}
	}
	return
}

func (s *GRPCServer) GetOwnUrlList(c context.Context, g *pb.GetOwnUrlListRequest) (r *pb.GetOwnUrlListResponse, err error) {
	ownerID, token := getOwnerAndToken(c)

	r = &pb.GetOwnUrlListResponse{}
	r.R = &pb.Resp{Code: http.StatusOK, Msg: "", Token: token}

	var (
		urls []dbf.StorageURL
		e    error
	)

	urls, e = dbf.Store.DBFGetOwnURLs(c, ownerID)

	if e != nil {
		r.R.Code = http.StatusInternalServerError
		r.R.Msg = e.Error()
		return
	}

	for _, url := range urls {
		if url.OriginalURL != "" {
			link := &pb.Curl{}
			link.CorrelationId = url.ShortURL
			link.Url = url.OriginalURL
			r.Urls = append(r.Urls, link)
		}
	}

	if len(r.Urls) == 0 {
		r.R.Code = http.StatusNoContent
		return
	}
	return
}

func (s *GRPCServer) GetStats(c context.Context, g *pb.GetStatsRequest) (r *pb.GetStatsResponse, err error) {
	r = &pb.GetStatsResponse{}
	r.R = &pb.Resp{Code: http.StatusOK, Msg: ""}

	fb := getForbidden(c)
	if fb {
		r.R.Code = http.StatusForbidden
		r.R.Msg = "Запрещено"
		return
	}

	stat, err := dbf.Store.GetStats(c)

	if err != nil {
		r.R.Code = http.StatusInternalServerError
		r.R.Msg = err.Error()
		return
	}

	r.Surls = int32(stat.URLs)
	r.Users = int32(stat.Users)
	return
}
