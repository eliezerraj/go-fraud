package service

import(
	"context"
	"strconv"
	"fmt"
	"math"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sagemakerruntime"

	"github.com/go-fraud/internal/core/model"
	go_core_observ "github.com/eliezerraj/go-core/observability"
)

var tracerProvider go_core_observ.TracerProvider

type Point struct {
    X float64
    Y float64
}

// About invoke fraud score model
func (s *WorkerService) CheckPaymentFraud(	ctx context.Context, 
											awsRegion string,
										 	payment model.PaymentFraud) (*model.PaymentFraud, error){
	childLogger.Info().Str("func","CheckPaymentFraud").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("payment", payment).Send()

	// Trace
	span := tracerProvider.Span(ctx, "service.CheckPaymentFraud")
	defer span.End()
	
	// Load aws config
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
    if err != nil {
        childLogger.Error().Err(err).Send()
        return nil, err
    }

	// create sagemaker client
	client := sagemakerruntime.NewFromConfig(cfg)

	// business rule
	var ohe_card_model_chip, ohe_card_model_virtual, ohe_card_type int
	if payment.CardModel == "VIRTUAL" {
		ohe_card_model_chip = 0
		ohe_card_model_virtual = 1
	} else {
		ohe_card_model_chip = 1
		ohe_card_model_virtual = 0
	}

	// prepare features
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
	
	childLogger.Info().Msg("header:distance,ohe_card_model_chip,ohe_card_model_virtual,ohe_card_type,payment.Amount,payment.Tx1Day,payment.Avg1Day,payment.Tx7Day,payment.Avg7Day,payment.Tx30Day,payment.Avg30Day,payment.TimeBtwTx")
	childLogger.Info().Interface("payload", payload).Send()

	// prepare and call sagemaker
	spanChildSagemaker := tracerProvider.Span(ctx, "service.InvokeEndpoint")
	input := &sagemakerruntime.InvokeEndpointInput{EndpointName: &s.apiService[0].Url,
													ContentType:  aws.String("text/csv"),
													Body:         []byte(payload),
												}
	

	resp, err := client.InvokeEndpoint(ctx, input)
	if err != nil {
		childLogger.Error().Err(err).Send()
		return nil, err
	}
	defer spanChildSagemaker.End()		

	// handle response
	responseBody := string(resp.Body)
	
	responseFloat, err := strconv.ParseFloat(responseBody, 64)
	if err != nil {
		childLogger.Error().Err(err).Send()
		return nil, err
	}

	payment.Fraud = responseFloat

	childLogger.Info().Interface("payment.Fraud", payment.Fraud).Send()
								
	return &payment, nil
}