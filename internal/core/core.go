package core

type AppServer struct {
	InfoPod			*InfoPod 		`json:"info_pod"`
	InfoServer     	*InfoServer     `json:"info_server"`
}

type InfoPod struct {
	PodName				string `json:"pod_name"`
	OSPID				string `json:"os_pid"`
	IPAddress			string `json:"ip_address"`
	AvailabilityZone 	string `json:"availability_zone"`
	GrpcHost			string `json:"grpc_host"`
	OtelExportEndpoint	string `json:"otel_export_endpoint"`
	ConfigOTEL			*ConfigOTEL 
}

type InfoServer struct {
	Port 				string
	MaxConnectionAge	int
	MaxConnectionAgeGrace int
}