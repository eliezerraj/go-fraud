package main

import(
	"context"
	"net"
	"os"
	"io/ioutil"

	"github.com/joho/godotenv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/go-fraud/internal/core"
	"github.com/go-fraud/internal/handler"
	"github.com/go-fraud/internal/service"
)

var(
	logLevel = zerolog.DebugLevel
	infoPod		core.InfoPod
	infoServer	core.InfoServer
	appServer	core.AppServer
	configOTEL	core.ConfigOTEL
	sageMakerEndpoint string
	cert		core.Cert
	certPEM, certPrivKeyPEM	[]byte
	isTLS		= 	false
)

func getEnv() {
	log.Debug().Msg("getEnv")

	if os.Getenv("GRPC_HOST") !=  "" {
		infoPod.GrpcHost = os.Getenv("GRPC_HOST")
		infoServer.Port  = os.Getenv("GRPC_HOST")
	}
	if os.Getenv("POD_NAME") !=  "" {
		infoPod.PodName = os.Getenv("POD_NAME")
	}
	if os.Getenv("VERSION") !=  "" {
		infoPod.Version = os.Getenv("VERSION")
	}
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") !=  "" {	
		infoPod.OtelExportEndpoint = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	}

	if (os.Getenv("TLS") != "") {
		if (os.Getenv("TLS") != "false") {	
			isTLS = true
		}
	}

	if os.Getenv("SAGEMAKER_ENDPOINT") !=  "" {	
		sageMakerEndpoint = os.Getenv("SAGEMAKER_ENDPOINT")
	}
}

func init(){
	log.Debug().Msg("init")
	zerolog.SetGlobalLevel(logLevel)

	err := godotenv.Load(".env")
	if err != nil {
		log.Info().Err(err).Msg("No .env file !!!")
	}

	getEnv()

	configOTEL.TimeInterval = 1
	configOTEL.TimeAliveIncrementer = 1
	configOTEL.TotalHeapSizeUpperBound = 100
	configOTEL.ThreadsActiveUpperBound = 10
	configOTEL.CpuUsageUpperBound = 100
	configOTEL.SampleAppPorts = []string{}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Error().Err(err).Msg("Error to get the POD IP address !!!")
		os.Exit(3)
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				infoPod.IPAddress = ipnet.IP.String()
			}
		}
	}

	if (isTLS) {
		certPEM, err = ioutil.ReadFile("/var/pod/cert/serverB64.crt") // local: server_fraud_B64.crt
		if err != nil {
			log.Info().Err(err).Msg("Cert certPEM nao encontrado")
		} else {
			cert.CertPEM = certPEM
		}

		certPrivKeyPEM, err = ioutil.ReadFile("/var/pod/cert/serverB64.key") // local: server_fraud_B64.key
		if err != nil {
			log.Info().Err(err).Msg("Cert CertPrivKeyPEM nao encontrado")
		} else {
			cert.CertPrivKeyPEM = certPrivKeyPEM
		}
	}

	infoPod.ConfigOTEL = &configOTEL
	appServer.InfoPod =	&infoPod
	appServer.InfoServer = &infoServer
	appServer.Cert = &cert
}

func main() {
	log.Debug().Msg("----------------------------------------------------")
	log.Debug().Msg("main")
	log.Debug().Msg("----------------------------------------------------")
	log.Debug().Interface("=>AppServer : ", AppServer).Msg("")
	log.Debug().Msg("----------------------------------------------------")

	ctx := context.Background()

	service := service.NewWorkerService(sageMakerEndpoint)

	appGrpcServer := handler.NewAppGrpcServer(&appServer, service)
	appGrpcServer.StartGrpcServer(ctx)					
}