package lib

import (
	"time"
)

type AgentDetails struct {
	Name string `json:"name"`
	// ActiveServices *[]string `json:"services"`
	ActiveSince   *time.Time
	ConfigUpdated *time.Time
	// RoutingKeyOutput string
	RoutingKey string
	// QueueName    string
	// AgentDetails MicroService
}
