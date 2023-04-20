package lib

type MicroServiceData struct {
	Services map[string]MicroService `yaml:"services"`
}

// Map of network
//
//	for key, value := range v.Networks {
//		fmt.Println("network: " + key)
//		fmt.Println("list of aliases: " + strings.Join(value.Aliases, ","))
//	}
//
// Result would be:
// network: core_network and list of aliases: unl1_agent
// network: unl_1 list of aliases: unl1_agent
type MicroService struct {
	Tag         string
	Image       string
	Ports       map[string]string
	EnvVars     map[string]string
	Networks    map[string]Network
	NetworkList []string
	Secrets     []string
	Volumes     map[string]string
	Deploy      Deploy
}

type Network struct {
	Aliases []string
}

type Deploy struct {
	Replicas  int       `json:"replicas,omitempty" yaml:"replicas,omitempty"`
	Placement Placement `json:"placement,omitempty" yaml:"placement,omitempty"`
	Resources Resources `json:"resources,omitempty" yaml:"resources,omitempty"`
}

type Placement struct {
	Constraints []string `json:"constraints,omitempty" yaml:"constraints,omitempty"`
}

type Resources struct {
	Reservations Resource `json:"reservations,omitempty" yaml:"reservations,omitempty"`
	Limits       Resource `json:"limits,omitempty" yaml:"limits,omitempty"`
}

type Resource struct {
	Memory string `json:"memory,omitempty" yaml:"memory,omitempty"`
}

type ExternalDockerConfig struct {
	Networks []string `yaml:"networks"`
	Volumes  []string `yaml:"volumes"`
	Secrets  []string `yaml:"secrets"`
}

// type Service struct {
// 	Services map[string]CreateServicePayload `yaml:"services"`
// }
// type CreateServicePayload struct {
// 	ImageName string            `json:"image" yaml:"image"`
// 	Tag       string            `json:"tag,omitempty" yaml:"tag,omitempty"`
// 	EnvVars   map[string]string `json:"env_vars" yaml:"environment"`
// 	Networks  []string          `json:"networks" yaml:"networks"`
// 	Secrets   []string          `json:"secrets" yaml:"secrets"`
// 	Volumes   map[string]string `json:"volumes" yaml:"-"`
// 	Ports     map[string]string `json:"ports,omitempty" yaml:"-"`
// 	Deploy    Deploy            `json:"deploy,omitempty" yaml:"deploy"`
// }

// func (c CreateServicePayload) String() string {
// 	var sb strings.Builder

// 	sb.WriteString(fmt.Sprintf("ImageName: %s\n", c.ImageName))
// 	sb.WriteString(fmt.Sprintf("Tag: %s\n", c.Tag))
// 	sb.WriteString("EnvVars:\n")
// 	for k, v := range c.EnvVars {
// 		sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
// 	}
// 	sb.WriteString(fmt.Sprintf("Networks: %v\n", c.Networks))
// 	sb.WriteString(fmt.Sprintf("Secrets: %v\n", c.Secrets))
// 	sb.WriteString("Volumes:\n")
// 	for k, v := range c.Volumes {
// 		sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
// 	}
// 	sb.WriteString("Ports:\n")
// 	for k, v := range c.Ports {
// 		sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
// 	}
// 	sb.WriteString(fmt.Sprintf("Deploy: \n"))
// 	sb.WriteString(fmt.Sprintf("  Replicas: %d\n", c.Deploy.Replicas))
// 	sb.WriteString(fmt.Sprintf("  Placement: \n"))
// 	sb.WriteString(fmt.Sprintf("    Constraints: %v\n", c.Deploy.Placement.Constraints))
// 	sb.WriteString(fmt.Sprintf("  Resources: \n"))
// 	sb.WriteString(fmt.Sprintf("    Reservations: \n"))
// 	sb.WriteString(fmt.Sprintf("      Memory: %s\n", c.Deploy.Resources.Reservations.Memory))
// 	sb.WriteString(fmt.Sprintf("    Limits: \n"))
// 	sb.WriteString(fmt.Sprintf("      Memory: %s\n", c.Deploy.Resources.Limits.Memory))

// 	return sb.String()
// }
