package service

import(
	"context"
	"strconv"
	"fmt"
	"math"

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

type Point struct {
    X float64
    Y float64
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

	var ohe_card_model_chip, ohe_card_model_virtual,ohe_card_type int
	if payment.CardModel == "VIRTUAL" {
		ohe_card_model_chip = 0
		ohe_card_model_virtual = 1
	} else {
		ohe_card_model_chip = 1
		ohe_card_model_virtual = 0
	}

	ohe_card_type = 1

    person 			:= Point{0, 0}
    terminal_order 	:= Point{float64(payment.CoordX), float64(payment.CoordY)}
	distance := math.Sqrt(math.Pow(terminal_order.X-person.X, 2) + math.Pow(terminal_order.Y-person.Y, 2))

	payload := fmt.Sprintf("%v, %v, %v, %v, %v, %v, %v, %v, %v, %v, %v, %v", 
									distance,
									ohe_card_model_chip,
									ohe_card_model_virtual,
									ohe_card_type,
									payment.Amount,
									payment.Tx1Day,
									payment.Avg1Day,
									payment.Tx7Day,
									payment.Avg7Day,
									payment.Tx30Day,
									payment.Avg30Day,
									payment.TimeBtwTx)
	
	fmt.Println("=================> (payload) :",payload)

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