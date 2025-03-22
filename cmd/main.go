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
	logLevel = 	zerolog.InfoLevel // zerolog.InfoLevel zerolog.DebugLevel
	appServer	model.AppServer
	childLogger = log.With().Str("component","go-fraud").Str("package", "main").Logger()
)

// About initialize the enviroment var
func init(){
	childLogger.Info().Str("func","init").Send()

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
	childLogger.Info().Str("func","main").Interface("appServer :",appServer).Send()

	ctx := context.Background()

	// create and wire
	workerService := service.NewWorkerService(appServer.ApiService)
	serviceGrpcServer := service_grpc_server.NewServiceGrpcServer(&appServer, workerService)
	workerServer := server.NewWorkerServer(serviceGrpcServer)

	// start grpc server
	workerServer.StartGrpcServer(ctx, &appServer)
}