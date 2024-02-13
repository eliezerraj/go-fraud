package healthcheck

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc/health/grpc_health_v1"
  
  )


type HealthChecker struct{}

var startTime = time.Now()

func (s *HealthChecker) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	log.Println("Inside Health check Request")
	//var currentTime = time.Now()
	var currentStatus = grpc_health_v1.HealthCheckResponse_SERVING
	// simulating unavailability ater two minutes
	/*if currentTime.Sub(startTime).Minutes() > 2 {
		currentStatus = grpc_health_v1.HealthCheckResponse_NOT_SERVING
	}*/
	health_check_response := &grpc_health_v1.HealthCheckResponse{
		Status: currentStatus,
	}
	return health_check_response, nil
}

func (s *HealthChecker) Watch(req *grpc_health_v1.HealthCheckRequest, server grpc_health_v1.Health_WatchServer) error {
	log.Println("Inside Health check Watch")
	//var currentTime = time.Now()
	var currentStatus = grpc_health_v1.HealthCheckResponse_SERVING
	// simulating unavailability ater two minutes
	/*if currentTime.Sub(startTime).Minutes() > 2 {
		currentStatus = grpc_health_v1.HealthCheckResponse_NOT_SERVING
	}*/
	health_check_response := &grpc_health_v1.HealthCheckResponse{
		Status: currentStatus,
	}
	return server.Send(health_check_response)
}

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{}
}