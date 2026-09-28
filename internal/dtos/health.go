package dtos

// HealthServices represents the operational status of sub-services.
type HealthServices struct {
	Database string `json:"database"`
	Redis    string `json:"redis"`
	S3       string `json:"s3"`
	RabbitMQ string `json:"rabbitmq,omitempty"`
	SMTP     string `json:"smtp,omitempty"`
}

// HealthData contains runtime health check details.
type HealthData struct {
	Version  string         `json:"version"`
	GitHash  string         `json:"git_hash"`
	Uptime   string         `json:"uptime"`
	Services HealthServices `json:"services"`
}
