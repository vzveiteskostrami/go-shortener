package lgrpc

import (
	"context"
	"net/http"

	pb "github.com/vzveiteskostrami/go-shortener/internal/proto"
)

func (s *GRPCServer) GetLink(c context.Context, g *pb.GetLinkRequest) (r *pb.GetLinkResponse, err error) {
	r = &pb.GetLinkResponse{}
	r.R = &pb.Resp{Code: http.StatusOK, Msg: ""}
	return
}

func (s *GRPCServer) GetOwnUrlList(c context.Context, g *pb.GetOwnUrlListRequest) (r *pb.GetOwnUrlListResponse, err error) {
	r = &pb.GetOwnUrlListResponse{}
	r.R = &pb.Resp{Code: http.StatusOK, Msg: ""}
	return
}

func (s *GRPCServer) GetStats(c context.Context, g *pb.GetStatsRequest) (r *pb.GetStatsResponse, err error) {
	r = &pb.GetStatsResponse{}
	r.R = &pb.Resp{Code: http.StatusOK, Msg: ""}
	return
}
