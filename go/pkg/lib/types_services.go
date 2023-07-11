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
	RoutingKey string `json:"routingKey"`
	Dns        string `json:"dns"`
	// QueueName    string
	// AgentDetails MicroService
}

type JobUserInfo struct {
	Name        string `json:"name"`
	RequestType string `json:"requestType"`
	JobName     string `json:"jobName"`
	ArcheType   string `json:"archetype"`
	Role        string `json:"role"`
}
