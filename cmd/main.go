package main

import(
	"context"
	
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/go-fraud/internal/infra/configuration"
	"github.com/go-fraud/internal/core/model"
	"github.com/go-fraud/internal/core/service"
	"github.com/go-fraud/internal/infra/server"
	service_grpc_server "github.com/go-fraud/internal/adapter/grpc/server"
)

var(
	logLevel = 	zerolog.DebugLevel
	appServer	model.AppServer
)

// About initialize the enviroment var
func init(){
	log.Debug().Msg("init")
	zerolog.SetGlobalLevel(logLevel)

	infoPod, server := configuration.GetInfoPod()
	configOTEL 		:= configuration.GetOtelEnv()
	apiService 	:= configuration.GetEndpointEnv() 
	awsService 	:= configuration.GetAwsServiceEnv() 

	appServer.InfoPod = &infoPod
	appServer.Server = &server
	appServer.ConfigOTEL = &configOTEL
	appServer.ApiService = apiService
	appServer.AwsService = &awsService
}

// About main
func main (){
	log.Debug().Msg("----------------------------------------------------")
	log.Debug().Msg("main")
	log.Debug().Msg("----------------------------------------------------")
	log.Debug().Interface("appServer :",appServer).Msg("")
	log.Debug().Msg("----------------------------------------------------")

	ctx := context.Background()

	// create and wire
	workerService := service.NewWorkerService(appServer.ApiService)
	serviceGrpcServer := service_grpc_server.NewServiceGrpcServer(&appServer, workerService)
	workerServer := server.NewWorkerServer(serviceGrpcServer)

	// start grpc server
	workerServer.StartGrpcServer(ctx, &appServer)
}