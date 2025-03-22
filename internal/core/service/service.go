package service

import(
	"github.com/go-fraud/internal/core/model"
	"github.com/rs/zerolog/log"
)

var childLogger = log.With().Str("component","go-fraud").Str("package","internal.core.service").Logger()

type WorkerService struct {
	apiService		[]model.ApiService
}

func NewWorkerService(apiService	[]model.ApiService) *WorkerService{
	childLogger.Info().Str("func","NewWorkerService").Send()

	return &WorkerService{
		apiService: apiService,
	}
}