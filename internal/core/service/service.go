package service

import(
	"github.com/go-fraud/internal/core/model"
	"github.com/rs/zerolog/log"
)

var childLogger = log.With().Str("core", "service").Logger()

type WorkerService struct {
	apiService		[]model.ApiService
}

func NewWorkerService(apiService	[]model.ApiService) *WorkerService{
	childLogger.Debug().Msg("NewWorkerService")

	return &WorkerService{
		apiService: apiService,
	}
}