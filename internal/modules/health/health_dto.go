package health

// Services represents the operational status of sub-services.
type Services struct {
	Database string `json:"database"`
	Redis    string `json:"redis"`
	S3       string `json:"s3"`
	RabbitMQ string `json:"rabbitmq,omitempty"`
	SMTP     string `json:"smtp,omitempty"`
}

// Data contains runtime health check details returned to the client.
type Data struct {
	Version  string   `json:"version"`
	GitHash  string   `json:"git_hash"`
	Uptime   string   `json:"uptime"`
	Services Services `json:"services"`
}
