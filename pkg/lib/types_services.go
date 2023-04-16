package lib

import (
	"time"
)

type AgentData struct {
	Agents map[string]AgentDetails `yaml:"services"`
}

type AgentDetails struct {
	Name             string    `json:"name"`
	ActiveServices   *[]string `json:"services"`
	ActiveSince      *time.Time
	ConfigUpdated    *time.Time
	RoutingKeyOutput string
	RoutingKeyInput  string
	InputQueueName   string
	ServiceName      string
	AgentDetails     MicroServiceDetails
}
