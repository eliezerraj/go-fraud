package server

import (
	"time"
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/go-fraud/internal/core/model"
	"github.com/go-fraud/internal/core/erro"
	service_grpc_server "github.com/go-fraud/internal/adapter/grpc/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/health/grpc_health_v1"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	
	"github.com/go-fraud/internal/infra/healthcheck"
	proto "github.com/go-fraud/internal/adapter/grpc/proto"
	

	"google.golang.org/grpc/metadata"
)

var childLogger = log.With().Str("component","go-fraud").Str("package","internal.infra.server").Logger()

var serviceGrpcServer service_grpc_server.ServiceGrpcServer
var tracer trace.Tracer

type WorkerServer struct {
	serviceGrpcServer *service_grpc_server.ServiceGrpcServer
}

// About create worker server
func NewWorkerServer(serviceGrpcServer *service_grpc_server.ServiceGrpcServer) *WorkerServer {
	childLogger.Info().Str("func","NewWorkerServer").Send()

	return &WorkerServer{
		serviceGrpcServer: serviceGrpcServer,
	}
}

// About authentication intercetor
func authenticationInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	childLogger.Info().Str("func","authenticationInterceptor").Send()

	headers, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, erro.MissingData
	}
	if len(headers["authorization"]) == 0 {
	//	return nil, erro.InvalidToken
		childLogger.Info().Msg("erro.InvalidToken")
	}
	return handler(ctx, req)
}

// About start server
func (w *WorkerServer) StartGrpcServer(	ctx context.Context, 
										appServer *model.AppServer){
	childLogger.Info().Str("func","authenticationInterceptor").Send()

	//Otel
	traceExporter, err := otlptracegrpc.New(ctx, 
											otlptracegrpc.WithInsecure(),
											otlptracegrpc.WithEndpoint(appServer.ConfigOTEL.OtelExportEndpoint),
											)
	if err != nil {
		childLogger.Error().Err(err).Msg("erro otlptracegrpc")
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
	tracer = otel.Tracer(appServer.InfoPod.PodName)

	// create grpc listener
	listener, err := net.Listen("tcp", appServer.Server.Port)
	if err != nil {
		childLogger.Error().Err(err).Msg("ERRO FATAL na abertura do service grpc")
		panic(err)
	}

	// prepare the options
	var opts []grpc.ServerOption
	opts = append(opts, grpc.ChainUnaryInterceptor( otelgrpc.UnaryServerInterceptor( ), authenticationInterceptor))
	opts = append(opts, grpc.KeepaliveParams(	keepalive.ServerParameters {
												MaxConnectionAge: time.Second * 30,
												MaxConnectionAgeGrace: time.Second * 10,
											}))
	
	// setup and prepare grpc server
	workerGrpcServer := grpc.NewServer(opts...)

	// handle defer
	defer func() {
		err = tp.Shutdown(ctx)
		if err != nil{
			childLogger.Error().Err(err).Send()
		}

		childLogger.Info().Msg("stopping server...")
		workerGrpcServer.Stop()
	
		childLogger.Info().Msg("stopping listener...")
		listener.Close()

		childLogger.Info().Msg("server stoped !!!")
	}()

	// wire
	proto.RegisterFraudServiceServer(workerGrpcServer, w.serviceGrpcServer)

	// health check
	healthService := healthcheck.NewHealthChecker()
	grpc_health_v1.RegisterHealthServer(workerGrpcServer, healthService)

	// run server
	go func(){
		childLogger.Info().Str("Starting server : ", appServer.Server.Port).Msg("")
		if err := workerGrpcServer.Serve(listener); err != nil {
			childLogger.Error().Err(err).Msg("Failed to server")
		}
	}()

	// get signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM )
	<-ch
}