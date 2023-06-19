package lib

import (
	"strings"
)

// Unmarshal a docker stack file into a struct of type YamlConfig.
// Will contain a list of all externally declared networks, secrets and volumes
func (c *ExternalDockerConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Define an intermediate structure to store the YAML data
	temp := struct {
		Networks map[string]struct {
			External bool `yaml:"external"`
		} `yaml:"networks"`
		Volumes map[string]struct {
			External bool `yaml:"external"`
		} `yaml:"volumes"`
		Secrets map[string]struct {
			External bool `yaml:"external"`
		} `yaml:"secrets"`
	}{}

	err := unmarshal(&temp)
	if err != nil {
		logger.Sugar().Errorw("Failed to marshal the payload to JSON: %v", err)
		return err
	}

	// Extract the names and store them in the respective fields
	for networkName := range temp.Networks {
		c.Networks = append(c.Networks, networkName)
	}

	for volumeName := range temp.Volumes {
		c.Volumes = append(c.Volumes, volumeName)
	}

	for secretName := range temp.Secrets {
		c.Secrets = append(c.Secrets, secretName)
	}

	return nil
}

func (ms *MicroServiceData) UnmarshalYAML(unmarshal func(interface{}) error) error {

	temp := struct {
		Services map[string]struct {
			Image    string            `yaml:"image"`
			EnvVars  map[string]string `yaml:"environment"`
			Networks interface{}       `yaml:"networks"`
			Secrets  []string          `yaml:"secrets"`
			Volumes  []string          `yaml:"volumes"`
			Ports    []string          `yaml:"ports,omitempty"`
			Deploy   Deploy            `yaml:"deploy"`
		} `yaml:"services"`
	}{}

	err := unmarshal(&temp)
	if err != nil {
		logger.Sugar().Errorw("Failed to unmarshal temp struct: %v", err)

		return err
	}

	ms.Services = make(map[string]MicroService)

	for serviceName, serviceDetails := range temp.Services {

		imageName, tag := SplitImageAndTag(serviceDetails.Image)

		volumes := make(map[string]string)
		for _, volume := range serviceDetails.Volumes {
			parts := strings.Split(volume, ":")
			if len(parts) == 2 {
				volumes[parts[0]] = parts[1]
			}
		}

		ports := make(map[string]string)
		for _, port := range serviceDetails.Ports {
			parts := strings.Split(port, ":")
			if len(parts) == 2 {
				ports[parts[0]] = parts[1]
			}
		}

		networks, networkMap := parseNetwork(serviceDetails.Networks)

		payload := MicroService{
			Image:       imageName,
			Tag:         tag,
			EnvVars:     serviceDetails.EnvVars,
			Secrets:     serviceDetails.Secrets,
			Networks:    networkMap,
			NetworkList: networks,
			Volumes:     volumes,
			Ports:       ports,
			Deploy:      serviceDetails.Deploy,
		}

		ms.Services[serviceName] = payload
	}

	return nil
}

func parseNetwork(networkInterface interface{}) ([]string, map[string]Network) {

	var networks []string
	var networkMap map[string]Network

	switch v := networkInterface.(type) {
	case []interface{}:
		for _, network := range v {
			networks = append(networks, network.(string))
		}
	case map[interface{}]interface{}:
		networkMap = make(map[string]Network)
		for k, v := range v {
			networkName := k.(string)
			networkData := v.(map[interface{}]interface{})
			aliases := make([]string, 0)
			if aliasList, ok := networkData["aliases"]; ok {
				for _, alias := range aliasList.([]interface{}) {
					aliases = append(aliases, alias.(string))
				}
			}
			networkMap[networkName] = Network{Aliases: aliases}
		}
	default:
		logger.Sugar().Errorw("Unsupported network type")
		return networks, networkMap
	}

	return networks, networkMap
}
