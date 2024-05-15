package util

import(
	"os"
	"strconv"
	"net"
	"context"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/go-fraud/internal/core"
)

var childLogger = log.With().Str("util", "util").Logger()

func GetInfoPod() (	core.InfoPod,
					core.Server, 
					core.ServiceEndpoint) {
	childLogger.Debug().Msg("GetInfoPod")

	err := godotenv.Load(".env")
	if err != nil {
		childLogger.Info().Err(err).Msg("No .env File !!!!")
	}

	var infoPod 	core.InfoPod
	var server		core.Server
	var serviceEndpoint core.ServiceEndpoint

	server.MaxConnectionAge = 30
	server.MaxConnectionAgeGrace = 30

	if os.Getenv("API_VERSION") !=  "" {
		infoPod.ApiVersion = os.Getenv("API_VERSION")
	}
	if os.Getenv("PORT") !=  "" {
		server.Port = os.Getenv("PORT")
	}
	if os.Getenv("POD_NAME") !=  "" {
		infoPod.PodName = os.Getenv("POD_NAME")
	}
	if os.Getenv("SETPOD_AZ") == "false" {	
		infoPod.IsAZ = false
	} else {
		infoPod.IsAZ = true
	}

	if os.Getenv("SAGEMAKER_ENDPOINT") !=  "" {
		serviceEndpoint.ServiceUrlDomain = os.Getenv("SAGEMAKER_ENDPOINT")
	}

	// Get IP
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
	infoPod.OSPID = strconv.Itoa(os.Getpid())

	// Get AZ only if localtest is true
	if (infoPod.IsAZ == true) {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			childLogger.Error().Err(err).Msg("ERRO FATAL get Context !!!")
			os.Exit(3)
		}
		client := imds.NewFromConfig(cfg)
		response, err := client.GetInstanceIdentityDocument(context.TODO(), &imds.GetInstanceIdentityDocumentInput{})
		if err != nil {
			childLogger.Error().Err(err).Msg("Unable to retrieve the region from the EC2 instance !!!")
			os.Exit(3)
		}
		infoPod.AvailabilityZone = response.AvailabilityZone	
	} else {
		infoPod.AvailabilityZone = "-"
	}

	return infoPod, server, serviceEndpoint
}
