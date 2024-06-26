package handler

import(
	"context"
	"fmt"
	"net"
	"os"
	"time"
	"os/signal"
	"syscall"
	"encoding/base64"
	"crypto/tls"
	
	"github.com/go-fraud/internal/core"
	"github.com/go-fraud/internal/service"

	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"github.com/go-fraud/internal/healthcheck"
	proto "github.com/go-fraud/internal/proto"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc/filters"
	
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var childLogger = log.With().Str("handler", "handler").Logger()
var tracer trace.Tracer
//-------------------------------------------------------
type server struct{
	AppServer			*core.AppServer
	Workerservice		*service.WorkerService
}
//------------------------------------------------------
type AppGrpcServer struct {
	appServer 			*core.AppServer
	workerservice		*service.WorkerService
}

func NewAppGrpcServer(	appServer *core.AppServer,
						workerservice	*service.WorkerService) (AppGrpcServer) {
	childLogger.Debug().Msg("NewAppGrpcServerr")

	return AppGrpcServer{
		appServer: appServer,
		workerservice: workerservice,
	}
}
//-------------------------------------------------------
func middleware(ctx context.Context, 
				req interface{}, 
				info *grpc.UnaryServerInfo, 
				handler grpc.UnaryHandler) (interface{}, error){
	childLogger.Debug().Msg("----------------------------------------------------")

	headers, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "INTERNAL_SERVER_ERROR")
	}
	
	if len(headers["authorization"]) == 0 {
		return nil, status.Error(codes.Unauthenticated, "Not Authorized")
	}

	childLogger.Debug().Msg("----------------------------------------------------")
	return handler(ctx, req) 
}

func (s server) GetPodInfo(	ctx context.Context, 
							in *proto.PodInfoRequest) (*proto.PodInfoResponse, error) {
	childLogger.Debug().Msg("GetPodInfo")

	tracer := otel.Tracer("go-fraud")
	_, span := tracer.Start(ctx, 
							"svc.GetPodInfo",
						)
	defer span.End()

	podInfo := proto.PodInfo{	IpAddress: 			s.AppServer.InfoPod.IPAddress,
								PodName: 			s.AppServer.InfoPod.PodName,
								AvailabilityZone:	s.AppServer.InfoPod.AvailabilityZone,
								GrpcHost:			s.AppServer.Server.Port,
								Version:			s.AppServer.InfoPod.ApiVersion,
							}

	res := &proto.PodInfoResponse {
		PodInfo: &podInfo,
	}

	childLogger.Debug().Interface("res :", res).Msg("")

	return res, nil
}

func (s server) CheckPaymentFraud(	ctx context.Context, 
									in *proto.PaymentRequest) (*proto.PaymentResponse, error) {
	log.Debug().Msg("CheckPaymentFraud")

	tracer := otel.Tracer("go-fraud")
	_, span := tracer.Start(ctx, 
							"svc.CheckPaymentFraud",
						)
	defer span.End()

	//log.Debug().Interface("in :", in).Msg("")
	log.Debug().Msg("---------------------------------------------")

	paymentAt := in.Payment.PaymentAt.AsTime()

	paymentFraud := core.PaymentFraud {	AccountID: in.Payment.AccountId,
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

	res_payment_fraud, err := s.Workerservice.CheckPaymentFraud(ctx, paymentFraud)								
	if err != nil {
		log.Error().Err(err).Msg("error CheckPaymentFraud")
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

func loadServerCertsTLS(cert *core.Cert) (credentials.TransportCredentials, error) {
	log.Debug().Msg("loadCertsTLS")

	var serverTLSConf *tls.Config

	certPEM_Raw, err := base64.StdEncoding.DecodeString(string(cert.CertPEM))
	if err != nil {
		childLogger.Error().Err(err).Msg("Erro certPEM_Raw !!!")
		return nil, err
	}
	certPrivKeyPEM_Raw, err := base64.StdEncoding.DecodeString(string(cert.CertPrivKeyPEM))
	if err != nil {
		childLogger.Error().Err(err).Msg("Erro certPrivKeyPEM_Raw !!!")
		return nil, err
	}
	
	childLogger.Info().Msg("-------------------Server CRT-----------------------------")
	fmt.Println(string(certPEM_Raw))
	childLogger.Info().Msg("--------------------Server Key --------------------")
	fmt.Println(string(certPrivKeyPEM_Raw))
	childLogger.Info().Msg("------------------------------------------------")

	serverCert, err := tls.X509KeyPair( certPEM_Raw, certPrivKeyPEM_Raw)
	if err != nil {
		childLogger.Error().Err(err).Msg("Erro Load X509 KeyPair")
		return nil, err
	}

	serverTLSConf = &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert,
	}

	return credentials.NewTLS(serverTLSConf), nil
}

func (g AppGrpcServer) StartGrpcServer(ctx context.Context){
	childLogger.Info().Msg("StartGrpcServer")
	
	// ---------------------- OTEL
	childLogger.Info().Str("OTEL_EXPORTER_OTLP_ENDPOINT :", g.appServer.ConfigOTEL.OtelExportEndpoint).Msg("")
	traceExporter, err := otlptracegrpc.New(ctx, 
											otlptracegrpc.WithInsecure(),
											otlptracegrpc.WithEndpoint(g.appServer.ConfigOTEL.OtelExportEndpoint),
											)
	if err != nil {
		log.Error().Err(err).Msg("ERRO otlptracegrpc")
	}
	idg := xray.NewIDGenerator()

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("go-fraud"),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithIDGenerator(idg),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(xray.Propagator{})

	defer func() { 
		err = tp.Shutdown(ctx)
		if err != nil{
			childLogger.Error().Err(err).Msg("Erro closing OTEL tracer !!!")
		}
	}()

	lis, err := net.Listen("tcp", g.appServer.Server.Port)
	if err != nil {
		log.Error().Err(err).Msg("ERRO FATAL na abertura do service grpc")
		panic(err)
	}

	var opts []grpc.ServerOption
	opts = append(opts, grpc.UnaryInterceptor( 	otelgrpc.UnaryServerInterceptor(
												otelgrpc.WithInterceptorFilter(filters.Not(filters.HealthCheck())),
											)))
	opts = append(opts, grpc.KeepaliveParams(	keepalive.ServerParameters {
												MaxConnectionAge: time.Second * 30,
												MaxConnectionAgeGrace: time.Second * 10,
											}))

	// ----------------------------------
	if g.appServer.Server.Cert.IsTLS == true  {
		tlsCredentials, err := loadServerCertsTLS(g.appServer.Server.Cert)
		if err != nil {
			log.Error().Err(err).Msg("Erro loadCertsTLS")
		}
		opts = append(opts, grpc.Creds( tlsCredentials ))
	}	
	//------------------------------------------

	_grpcServer := grpc.NewServer(opts...)

	_server := 	&server{ 	AppServer: 		g.appServer,
							Workerservice: 	g.workerservice }

  	proto.RegisterFraudServiceServer(_grpcServer, 
									_server)

	healthService := healthcheck.NewHealthChecker()

	grpc_health_v1.RegisterHealthServer(_grpcServer, 
										healthService)

	go func(){
		log.Info().Str("Starting server : " , g.appServer.Server.Port).Msg("")
		if err := _grpcServer.Serve(lis); err != nil {
			log.Error().Err(err).Msg("Failed to server")
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM )
	<-ch

	log.Info().Msg("Stopping server")
	_grpcServer.Stop()

	log.Info().Msg("Stopping listener")
	lis.Close()

	log.Info().Msg("Done !!")	
}