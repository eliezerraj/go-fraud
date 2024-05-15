package util

import(
	"os"
	"io/ioutil"

	"github.com/joho/godotenv"
	"github.com/go-fraud/internal/core"
)

func GetCertEnv() core.Cert {
	childLogger.Debug().Msg("GetCertEnv")

	err := godotenv.Load(".env")
	if err != nil {
		childLogger.Info().Err(err).Msg("No .env File !!!!")
	}

	var cert		core.Cert

	if os.Getenv("TLS") !=  "false" {	
		cert.IsTLS = true
		
		childLogger.Debug().Msg("*** Loading cert server_fraud_B64.crt ***")
		cert.CertPEM, err = ioutil.ReadFile("/var/pod/cert/server_fraud_B64.crt") // server_account_B64.crt
		if err != nil {
			childLogger.Info().Err(err).Msg("Cert certPEM nao encontrado")
		}
		
		childLogger.Debug().Msg("*** Loading cert server_fraud_B64.key ***")
		cert.CertPrivKeyPEM, err = ioutil.ReadFile("/var/pod/cert/server_fraud_B64.key") // server_account_B64.key
		if err != nil {
			childLogger.Info().Err(err).Msg("Cert CertPrivKeyPEM nao encontrado")
		}
	}

	return cert
}