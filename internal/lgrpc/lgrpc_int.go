package lgrpc

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/vzveiteskostrami/go-shortener/internal/auth"
	"github.com/vzveiteskostrami/go-shortener/internal/config"
	"github.com/vzveiteskostrami/go-shortener/internal/logging"
	"github.com/vzveiteskostrami/go-shortener/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

//const string showProhabited = "просмотр статистики запрещён"

func unaryInterceptorAuth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	needOwner := true
	switch req.(type) {
	case *proto.PingRequest:
		needOwner = false
	case *proto.GetLinkRequest:
		needOwner = false
	case *proto.GetStatsRequest:
		needOwner = false
	}

	if !needOwner {
		return handler(ctx, req)
	}

	var token string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("token")
		if len(values) > 0 {
			token = values[0]
		}
	}

	var ownerID int64 = 0
	var err error
	ownerValid := true

	if token == "" {
		ownerValid = false
		token, ownerID, err = auth.TokenService.MakeToken()
	} else if ownerID, err = auth.GetOwnerID(token); err != nil {
		ownerValid = false
		token, ownerID, err = auth.TokenService.MakeToken()
	}

	if err != nil {
		logging.S().Error(err)
		return nil, err
	}

	var v string
	if ownerValid {
		v = "t"
	} else {
		v = "f"
	}
	md := metadata.Pairs("owner", strconv.FormatInt(ownerID, 10), "valid", v, "token", token)

	c := metadata.NewIncomingContext(context.Background(), md)
	return handler(c, req)
}

func unaryInterceptorLog(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	a, err := handler(ctx, req)
	duration := time.Since(start)

	logging.S().Infoln(
		"method:", getFuncName(info.FullMethod),
		"duration:", duration,
	)

	return a, err
}

func unaryInterceptorTrust(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	needTrust := true
	switch req.(type) {
	case *proto.GetStatsRequest:
		needTrust = true
	default:
		needTrust = false
	}

	if !needTrust {
		return handler(ctx, req)
	}

	fb := "f"
	if config.Storage.TrustedIPNet == nil {
		fb = "t"
	} else {
		p, _ := peer.FromContext(ctx)
		ip := net.ParseIP(p.Addr.String())
		if ip == nil || !config.Storage.TrustedIPNet.Contains(ip) {
			fb = "t"
		}
	}

	md := metadata.Pairs("fb", fb)
	c := metadata.NewIncomingContext(context.Background(), md)
	return handler(c, req)
}
