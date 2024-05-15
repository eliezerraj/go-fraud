package main

import(
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/go-fraud/internal/util"
	"github.com/go-fraud/internal/core"
	"github.com/go-fraud/internal/handler"
	"github.com/go-fraud/internal/service"
)

var(
	logLevel = zerolog.DebugLevel
	appServer	core.AppServer
)

func init(){
	log.Debug().Msg("init")
	zerolog.SetGlobalLevel(logLevel)

	infoPod, server, serviceEndpoint := util.GetInfoPod()
	configOTEL := util.GetOtelEnv()
	cert := util.GetCertEnv()

	appServer.InfoPod = &infoPod
	appServer.Server = &server
	appServer.ServiceEndpoint = &serviceEndpoint
	appServer.Server.Cert = &cert
	appServer.ConfigOTEL = &configOTEL
}

func main() {
	log.Debug().Msg("----------------------------------------------------")
	log.Debug().Msg("main")
	log.Debug().Msg("----------------------------------------------------")
	log.Debug().Interface("=>AppServer : ", appServer).Msg("")
	log.Debug().Msg("----------------------------------------------------")

	ctx := context.Background()

	service := service.NewWorkerService(appServer.ServiceEndpoint.ServiceUrlDomain) //sageMakerEndpoint)

	appGrpcServer := handler.NewAppGrpcServer(	&appServer, 
												service)
	appGrpcServer.StartGrpcServer(ctx)					
}