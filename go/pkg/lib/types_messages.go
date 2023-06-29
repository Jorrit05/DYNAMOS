package lib

type OrchestratorRequest struct {
	Type      string   `json:"type"`
	Providers []string `json:"providers"`
	Query     string   `json:"query"`
	// Architecture IoConfig `json:"architecture"`
	Name string `json:"name"`
}

type DetachAttachServicePayload struct {
	ServiceName string `json:"service_name"`
	QueueName   string `json:"queue_name"`
}

type KillServicePayload struct {
	ServiceName string `json:"service_name"`
}

type ChangeReplicas struct {
	Type         string `json:"type"`
	ServiceName  string `json:"service_name"`
	NrOfReplicas uint64 `json:"replicas"`
}
