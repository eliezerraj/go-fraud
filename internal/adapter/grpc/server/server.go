package server

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/go-fraud/internal/core/service"
	"github.com/go-fraud/internal/core/model"

	proto "github.com/go-fraud/internal/adapter/grpc/proto"
	go_core_observ "github.com/eliezerraj/go-core/observability"
)

var childLogger = log.With().Str("adapter.grpc", "grpc_server").Logger()
var tracerProvider go_core_observ.TracerProvider

type ServiceGrpcServer struct{
	appServer *model.AppServer
	workerService *service.WorkerService
}

func NewServiceGrpcServer(	appServer *model.AppServer, 
							workerService *service.WorkerService) *ServiceGrpcServer {
	childLogger.Debug().Msg("NewServiceGrpcServer")

	return &ServiceGrpcServer{
		appServer: appServer,
		workerService: workerService,
	}
}

// About get pod information
func (s *ServiceGrpcServer) GetPodInfo(ctx context.Context, 
								in *proto.PodInfoRequest) (*proto.PodInfoResponse, error) {
	childLogger.Debug().Msg("GetPodInfo")

	// Trace
	span := tracerProvider.Span(ctx, "adpater.grpc.server.GetPodInfo")
	defer span.End()
	
	podInfo := proto.PodInfo{	IpAddress: 			s.appServer.InfoPod.IPAddress,
								PodName: 			s.appServer.InfoPod.PodName,
								AvailabilityZone:	s.appServer.InfoPod.AvailabilityZone,
								GrpcHost:			s.appServer.Server.Port,
								Version:			s.appServer.InfoPod.ApiVersion,
							}

	res := &proto.PodInfoResponse {
		PodInfo: &podInfo,
	}

	childLogger.Debug().Interface("res :", res).Msg("")

	return res, nil
}

// About invoke fraud score model
func (s *ServiceGrpcServer) CheckPaymentFraud(	ctx context.Context, 
												in *proto.PaymentRequest) (*proto.PaymentResponse, error) {
	childLogger.Debug().Msg("CheckPaymentFraud")

	// Trace
	span := tracerProvider.Span(ctx, "adpater.grpc.server.CheckPaymentFraud")
	defer span.End()

	paymentAt := in.Payment.PaymentAt.AsTime()

	paymentFraud := model.PaymentFraud {AccountID: in.Payment.AccountId,
										CardNumber: in.Payment.CardNumber,
										TerminalName: in.Payment.TerminalName, 
										CoordX: in.Payment.CoordX, 
										CoordY: in.Payment.CoordY, 
										PaymentAt: paymentAt, 
										CardType: in.Payment.CardType,
										CardModel: in.Payment.CardModel,
										Currency: in.Payment.Currency,
										MCC: in.Payment.Mcc,
										Amount: in.Payment.Amount,
										Status: in.Payment.Status,
										Tx1Day: in.Payment.Tx_1D,
										Avg1Day: in.Payment.Avg_1D,
										Tx7Day: in.Payment.Tx_7D,
										Avg7Day: in.Payment.Avg_7D,
										Tx30Day: in.Payment.Tx_30D,
										Avg30Day: in.Payment.Avg_30D,
										TimeBtwTx: in.Payment.TimeBtwCcTx,
									}

	res_payment_fraud, err := s.workerService.CheckPaymentFraud(ctx, s.appServer.AwsService.AwsRegion, paymentFraud)
	if err != nil {
		return nil, err
	}

	res_payment := proto.Payment{	AccountId: in.Payment.AccountId,
									CardNumber: in.Payment.CardNumber,
									TerminalName: in.Payment.TerminalName, 
									PaymentAt: in.Payment.PaymentAt, 
									CardType: in.Payment.CardType,
									CardModel: in.Payment.CardModel,
									Currency: in.Payment.Currency,
									Mcc: in.Payment.Mcc,
									Amount: in.Payment.Amount,
									Status: in.Payment.Status,
									Tx_1D: in.Payment.Tx_1D,
									Avg_1D: in.Payment.Avg_1D,
									Tx_7D: in.Payment.Tx_7D,
									Avg_7D: in.Payment.Avg_7D,
									Tx_30D: in.Payment.Tx_30D,
									Avg_30D: in.Payment.Avg_30D,
									TimeBtwCcTx: in.Payment.TimeBtwCcTx,
									Fraud: res_payment_fraud.Fraud,
								}

	res := &proto.PaymentResponse {
		Payment: &res_payment,
	}
	return res, nil
}