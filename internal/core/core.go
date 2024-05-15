package core

type AppServer struct {
	InfoPod 		*InfoPod 			`json:"info_pod"`
	Server     		*Server     		`json:"server"`
	ServiceEndpoint	*ServiceEndpoint	`json:"sagemaker_endpoint"`
	ConfigOTEL		*ConfigOTEL			`json:"otel_config"`
}

type InfoPod struct {
	PodName				string `json:"pod_name"`
	ApiVersion			string `json:"version"`
	OSPID				string `json:"os_pid"`
	IPAddress			string `json:"ip_address"`
	AvailabilityZone 	string `json:"availabilityZone"`
	IsAZ				bool	`json:"is_az"`
}

type ServiceEndpoint struct {
	ServiceUrlDomain 	string `json:"sagemaker_endpoint"`
}

type Server struct {
	Port 					string
	MaxConnectionAge		int
	MaxConnectionAgeGrace 	int
	Cert					*Cert	`json:"cert"`
}

type Cert struct {
	IsTLS				bool
	CertPEM 			[]byte 		
	CertPrivKeyPEM	    []byte     
}

type ConfigOTEL struct {
	OtelExportEndpoint		string
	TimeInterval            int64    `mapstructure:"TimeInterval"`
	TimeAliveIncrementer    int64    `mapstructure:"RandomTimeAliveIncrementer"`
	TotalHeapSizeUpperBound int64    `mapstructure:"RandomTotalHeapSizeUpperBound"`
	ThreadsActiveUpperBound int64    `mapstructure:"RandomThreadsActiveUpperBound"`
	CpuUsageUpperBound      int64    `mapstructure:"RandomCpuUsageUpperBound"`
	SampleAppPorts          []string `mapstructure:"SampleAppPorts"`
}