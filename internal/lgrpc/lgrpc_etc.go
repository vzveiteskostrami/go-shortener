package lgrpc

import (
	"context"
	"net/http"

	"github.com/vzveiteskostrami/go-shortener/internal/dbf"
	pb "github.com/vzveiteskostrami/go-shortener/internal/proto"
	"github.com/vzveiteskostrami/go-shortener/internal/urlman"
)

func (s *GRPCServer) DelOwnUrlList(c context.Context, d *pb.DelOwnUrlListRequest) (r *pb.DelOwnUrlListResponse, err error) {
	ownerID, token := getOwnerAndToken(c)

	r = &pb.DelOwnUrlListResponse{}
	r.R = &pb.Resp{Code: http.StatusOK, Msg: "", Token: token}

	go func() {
		surl := ""
		for _, data := range d.Urls {
			if url, err := dbf.Store.FindLink(c, data, true); err == nil {
				if !url.Deleted && url.OWNERID == ownerID {
					surl = data
				}
			}
			if surl != "" {
				urlman.WriteToDel(surl)
				surl = ""
			}
		}
	}()

	return
}

func (s *GRPCServer) Ping(c context.Context, p *pb.PingRequest) (r *pb.PingResponse, err error) {
	r = &pb.PingResponse{}
	r.R = &pb.Resp{Code: http.StatusOK, Msg: ""}
	code, err := dbf.Store.PingDBf()
	r.R.Code = int32(code)
	if err != nil {
		r.R.Msg = err.Error()
	}
	return
}
