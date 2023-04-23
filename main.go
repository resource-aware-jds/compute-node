package main

import (
	"fmt"
	"github.com/resource-aware-jds/common-go/logger"
	"github.com/resource-aware-jds/common-go/proto"
	"github.com/resource-aware-jds/compute-node/config"
	"github.com/resource-aware-jds/compute-node/handler"
	"github.com/resource-aware-jds/compute-node/service"
	"google.golang.org/grpc"
	"log"
	"net"
)

var appConfig config.Config

func init() {
	appConfig = config.Load()
	logger.InitLogger(logger.Config{
		Env: appConfig.Env,
	})
}

func main() {

	//Job service
	jobService := service.NewJobService(appConfig)

	//Handler
	jobHandler := handler.NewJobGrpcServer(jobService)

	// GRPC
	lis, err := net.Listen("tcp", fmt.Sprint(":", appConfig.GRPC_SERVER_PORT))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterComputeNodeServer(s, jobHandler)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
