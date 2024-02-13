package handler

import(
	"context"
	"net"
	"os"
	"time"
	"os/signal"
	"syscall"
	
	"github.com/go-fraud/internal/core"
	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"
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

	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var childLogger = log.With().Str("handler", "handler").Logger()
var tracer trace.Tracer

type AppGrpcServer struct {
	appServer 	*core.AppServer
}

type server struct{
	InfoPod		*core.InfoPod
}

func NewAppGrpcServer(appServer *core.AppServer) (AppGrpcServer) {
	childLogger.Debug().Msg("NewAppGrpcServerr")
	return AppGrpcServer{
		appServer: appServer,
	}
}

func middleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error){
	log.Debug().Msg("----------------------------------------------------")

	headers, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "INTERNAL_SERVER_ERROR")
	}
	
	if len(headers["authorization"]) == 0 {
		return nil, status.Error(codes.Unauthenticated, "Not Authorized")
	}

	log.Debug().Msg("----------------------------------------------------")
	return handler(ctx, req) 
}

func (s server) GetPodInfo(ctx context.Context, in *proto.PodInfoRequest) (*proto.PodInfoResponse, error) {
	log.Debug().Msg("GetPodInfo")

	tracer := otel.Tracer("go-fraud")
	_, span := tracer.Start(ctx, 
							"svc.GetPodInfo",
						)
	defer span.End()
	time.Sleep(1 * time.Second) 

	podInfo := proto.PodInfo{	IpAddress: 			s.InfoPod.IPAddress,
								PodName: 			s.InfoPod.PodName,
								AvailabilityZone:	s.InfoPod.AvailabilityZone,
								GrpcHost:			s.InfoPod.GrpcHost,
							}

	res := &proto.PodInfoResponse {
		PodInfo: &podInfo,
	}

	log.Debug().Interface("res :", res).Msg("")

	return res, nil
}

func (g AppGrpcServer) StartGrpcServer(ctx context.Context){
	childLogger.Info().Msg("StartGrpcServer")
	
	// ---------------------- OTEL
	childLogger.Info().Str("OTEL_EXPORTER_OTLP_ENDPOINT :", g.appServer.InfoPod.OtelExportEndpoint).Msg("")
	traceExporter, err := otlptracegrpc.New(ctx, 
											otlptracegrpc.WithInsecure(),
											otlptracegrpc.WithEndpoint(g.appServer.InfoPod.OtelExportEndpoint),
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
	// ----------------------------------

	lis, err := net.Listen("tcp", g.appServer.InfoServer.Port)
	if err != nil {
		log.Error().Err(err).Msg("ERRO FATAL na abertura do service grpc")
		panic(err)
	}

	var opts []grpc.ServerOption
	opts = append(opts, grpc.UnaryInterceptor( otelgrpc.UnaryServerInterceptor() ))
	opts = append(opts, grpc.KeepaliveParams(keepalive.ServerParameters {
															MaxConnectionAge: time.Second * 30,
															MaxConnectionAgeGrace: time.Second * 10,
											}))

	_server := 	&server{InfoPod: g.appServer.InfoPod}									
  	_grpcServer := grpc.NewServer(opts...)
  	proto.RegisterFraudServiceServer(_grpcServer, _server)

	healthService := healthcheck.NewHealthChecker()
	grpc_health_v1.RegisterHealthServer(_grpcServer, healthService)

	go func(){
		log.Info().Str("Starting server : " , g.appServer.InfoServer.Port).Msg("")
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