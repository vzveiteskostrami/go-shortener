package lgrpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/vzveiteskostrami/go-shortener/internal/config"
	"github.com/vzveiteskostrami/go-shortener/internal/logging"
	pb "github.com/vzveiteskostrami/go-shortener/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	pb.UnimplementedShortUrlServer
}

func DogRPC() {
	listen, err := net.Listen("tcp", ":"+strconv.Itoa(config.Addresses.In.Port))
	if err != nil {
		log.Fatal(err)
	}
	// создаём gRPC-сервер с middleware на логирование, аутентификацию и траст
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(unaryInterceptorLog, unaryInterceptorAuth, unaryInterceptorTrust))
	reflection.Register(s)
	// регистрируем сервис
	pb.RegisterShortUrlServer(s, &GRPCServer{})

	fmt.Println("Сервер gRPC начал работу. Порт ", config.Addresses.In.Port, ".")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)

	go func() {
		<-sigs
		if err := listen.Close(); err != nil {
			logging.S().Errorln("Server shutdown error", err)
		} else {
			logging.S().Infoln("Server has been closed succesfully")
		}
	}()

	// получаем запрос gRPC
	if err := s.Serve(listen); err != nil {
		if !strings.Contains(err.Error(), "use of closed network connection") {
			logging.S().Panic(err)
		}
	}
	logging.S().Infoln("This is the end")
}

func getFuncName(path string) string {
	s := strings.Split(path, "/")
	if len(s) > 0 {
		return s[len(s)-1]
	} else {
		return ""
	}
}

func getOwnerAndToken(ctx context.Context) (int64, string) {
	ownerID := int64(0)
	token := ""
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("owner")
		if len(values) > 0 {
			ownerID, _ = strconv.ParseInt(values[0], 10, 64)
		}
		values = md.Get("token")
		if len(values) > 0 {
			token = values[0]
		}
	}
	return ownerID, token
}

func getForbidden(ctx context.Context) bool {
	forbidden := true
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("fb")
		if len(values) > 0 {
			forbidden = values[0] == "t"
		}
	}
	return forbidden
}
