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
		log.Errorf("Failed to marshal the payload to JSON: %v", err)
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
			Image    string             `yaml:"image"`
			EnvVars  map[string]string  `yaml:"environment"`
			Networks map[string]Network `yaml:"networks"`
			Secrets  []string           `yaml:"secrets"`
			Volumes  []string           `yaml:"volumes"`
			Ports    []string           `yaml:"ports,omitempty"`
			Deploy   Deploy             `yaml:"deploy"`
		} `yaml:"services"`
	}{}

	err := unmarshal(&temp)
	if err != nil {
		log.Errorf("Failed to unmarshal temp struct: %v", err)

		return err
	}

	ms.Services = make(map[string]MicroServiceDetails)

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

		payload := MicroServiceDetails{
			Image:    imageName,
			Tag:      tag,
			EnvVars:  serviceDetails.EnvVars,
			Secrets:  serviceDetails.Secrets,
			Networks: serviceDetails.Networks,
			Volumes:  volumes,
			Ports:    ports,
			Deploy:   serviceDetails.Deploy,
		}

		ms.Services[serviceName] = payload
	}

	return nil
}
