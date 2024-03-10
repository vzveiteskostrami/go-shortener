package lgrpc

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/vzveiteskostrami/go-shortener/internal/auth"
	"github.com/vzveiteskostrami/go-shortener/internal/logging"
	"github.com/vzveiteskostrami/go-shortener/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func unaryInterceptorAuth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var token string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("token")
		if len(values) > 0 {
			token = values[0]
		}
	}

	fmt.Printf("Type=")
	switch tp := req.(type) {
	case *proto.SetLinkRequest:
		fmt.Println("GetLink")
	case bool:
		fmt.Println("boolean")
	default:
		fmt.Printf("%T/n", tp)
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

	/*
		logging.S().Infoln(
			"uri:", r.RequestURI,
			"method:", r.Method,
			"status:", responseData.status, http.StatusText(responseData.status),
			"duration:", duration,
			"size:", responseData.size,
		)
	*/

	return a, err

}

func unaryInterceptorTrust(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	/*
		var token string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			values := md.Get("token")
			if len(values) > 0 {
				token = values[0]
			}
		}
		if len(token) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing token")
		}
		if token != "asdf" {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
	*/
	return handler(ctx, req)
}
