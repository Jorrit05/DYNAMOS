package lib

type OrchestratorRequest struct {
	Type         string   `json:"type"`
	Providers    []string `json:"providers"`
	Query        string   `json:"query"`
	Architecture string   `json:"architecture"`
}

type DetachAttachServicePayload struct {
	ServiceName string `json:"service_name"`
	QueueName   string `json:"queue_name"`
}

type KillServicePayload struct {
	ServiceName string `json:"service_name"`
}
