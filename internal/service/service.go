package service

import(
	"context"
	"strconv"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sagemakerruntime"

	"github.com/go-fraud/internal/core"
	"go.opentelemetry.io/otel"
)


var childLogger = log.With().Str("service", "service").Logger()

type WorkerService struct {
	sageMakerEndpoint	string
}

func NewWorkerService(sageMakerEndpoint	string) *WorkerService{
	childLogger.Debug().Msg("NewWorkerService")

	return &WorkerService{
		sageMakerEndpoint:	sageMakerEndpoint,
	}
}

func (s WorkerService) CheckPaymentFraud(ctx context.Context, 
										 payment core.PaymentFraud) (*core.PaymentFraud, error){
	childLogger.Debug().Msg("CheckPaymentFraud")
	log.Debug().Interface("=======>payment :", payment).Msg("")

	ctx, svcspan := otel.Tracer("go-fraud").Start(ctx,"svc.CheckPaymentFraud")
	defer svcspan.End()

	cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        childLogger.Error().Err(err).Msg("error LoadDefaultConfig")
        return nil, err
    }

	client := sagemakerruntime.NewFromConfig(cfg)

	childLogger.Debug().Msg("*********************************************")
	fmt.Printf("%v, %v, %v, %v, %v, %v, %v, %v", 
									payment.Amount,
									payment.Tx1Day,
									payment.Avg1Day,
									payment.Tx7Day,
									payment.Avg7Day,
									payment.Tx30Day,
									payment.Avg30Day,
									payment.TimeBtwTx)
	
	payload := "9.0, 23.0, 7.0, 90.0, 4.0, 365.0, 17.0, 263.529412, 28.0, 238.714286, 97582.0"
	input := &sagemakerruntime.InvokeEndpointInput{EndpointName: &s.sageMakerEndpoint,
													ContentType:  aws.String("text/csv"),
													Body:         []byte(payload),
												}
	
	resp, err := client.InvokeEndpoint(ctx, input)
	if err != nil {
		childLogger.Error().Err(err).Msg("error InvokeEndpoint")
		return nil, err
	}
	
	responseBody := string(resp.Body)
	
	responseFloat, err := strconv.ParseFloat(responseBody, 64)
	if err != nil {
		childLogger.Error().Err(err).Msg("error ParseFloat")
		return nil, err
	}

	payment.Fraud = responseFloat

	return &payment, nil
}