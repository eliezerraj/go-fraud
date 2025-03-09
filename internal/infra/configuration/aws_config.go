package configuration

import(
	"os"

	"github.com/joho/godotenv"
	"github.com/go-fraud/internal/core/model"
)

func GetAwsServiceEnv() model.AwsService {
	childLogger.Debug().Msg("GetAwsServiceEnv")

	err := godotenv.Load(".env")
	if err != nil {
		childLogger.Info().Err(err).Msg("env file not found !!!")
	}
	
	var awsService	model.AwsService

	if os.Getenv("AWS_REGION") !=  "" {
		awsService.AwsRegion = os.Getenv("AWS_REGION")
	}

	return awsService
}