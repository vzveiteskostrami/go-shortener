package lgrpc

import (
	"context"
	"net/http"

	pb "github.com/vzveiteskostrami/go-shortener/internal/proto"
)

func (s *GRPCServer) DelOwnUrlList(c context.Context, d *pb.DelOwnUrlListRequest) (r *pb.DelOwnUrlListResponse, err error) {
	r = &pb.DelOwnUrlListResponse{}
	r.R = &pb.Resp{Code: http.StatusOK, Msg: ""}
	return
}

func (s *GRPCServer) Ping(c context.Context, p *pb.PingRequest) (r *pb.PingResponse, err error) {
	r = &pb.PingResponse{}
	r.R = &pb.Resp{Code: http.StatusOK, Msg: ""}

	return
}
