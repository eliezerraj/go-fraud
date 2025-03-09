package model

import (
	"time"
	go_core_observ "github.com/eliezerraj/go-core/observability" 
)

type AppServer struct {
	InfoPod 		*InfoPod 					`json:"info_pod"`
	Server     		*Server     				`json:"server"`
	ConfigOTEL		*go_core_observ.ConfigOTEL	`json:"otel_config"`
	AwsService		*AwsService					`json:"aws_service"`
	ApiService 		[]ApiService				`json:"api_endpoints"` 			
}

type InfoPod struct {
	PodName				string 	`json:"pod_name"`
	ApiVersion			string 	`json:"version"`
	OSPID				string 	`json:"os_pid"`
	IPAddress			string 	`json:"ip_address"`
	AvailabilityZone 	string 	`json:"availabilityZone"`
	IsAZ				bool   	`json:"is_az"`
	Env					string `json:"enviroment,omitempty"`
	AccountID			string `json:"account_id,omitempty"`
}

type Server struct {
	Port 			string `json:"port"`
	ReadTimeout		int `json:"readTimeout"`
	WriteTimeout	int `json:"writeTimeout"`
	IdleTimeout		int `json:"idleTimeout"`
	CtxTimeout		int `json:"ctxTimeout"`
}

type Account struct {
	ID				int			`json:"id,omitempty"`
	AccountID		string		`json:"account_id,omitempty"`
	PersonID		string  	`json:"person_id,omitempty"`
	CreateAt		time.Time 	`json:"create_at,omitempty"`
	UpdateAt		*time.Time 	`json:"update_at,omitempty"`
	TenantID		string  	`json:"tenant_id,omitempty"`
	UserLastUpdate	*string  	`json:"user_last_update,omitempty"`
}

type ApiService struct {
	Name			string `json:"name_service"`
	Url				string `json:"url"`
	Method			string `json:"method"`
	Header_x_apigw_api_id	string `json:"x-apigw-api-id"`
}

type AwsService struct {
	AwsRegion			string `json:"aws_region"`
}

type PaymentFraud struct {
	AccountID		string		`json:"account_id,omitempty"`
	CardNumber		string		`json:"card_number,omitempty"`
	TerminalName	string		`json:"terminal_name,omitempty"`
	CoordX			int32		`json:"coord_x,omitempty"`
	CoordY			int32		`json:"coord_y,omitempty"`
	CardType		string  	`json:"card_type,omitempty"`
	CardModel		string  	`json:"card_model,omitempty"`
	PaymentAt		time.Time	`json:"payment_at,omitempty"`
	MCC				string  	`json:"mcc,omitempty"`
	Status			string  	`json:"status,omitempty"`
	Currency		string  	`json:"currency,omitempty"`
	Amount			float64 	`json:"amount,omitempty"`
	Tx1Day			float64 	`json:"tx_1d,omitempty"`
	Avg1Day			float64 	`json:"avg_1d,omitempty"`
	Tx7Day			float64 	`json:"tx_7d,omitempty"`
	Avg7Day			float64 	`json:"avg_7d,omitempty"`
	Tx30Day			float64 	`json:"tx_30d,omitempty"`
	Avg30Day		float64 	`json:"avg_30d,omitempty"`
	TimeBtwTx		int32 		`json:"time_btw_cc_tx,omitempty"`
	Fraud			float64	  	`json:"fraud,omitempty"`
}
