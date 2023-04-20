package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

func GetDockerClient() *client.Client {
	// Create a new Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error creating Docker client: %v", err)
	}

	// Check if Swarm is active
	info, err := cli.Info(context.Background())
	if err != nil {
		log.Fatalf("Error getting Docker info: %v", err)
	}
	if !info.Swarm.ControlAvailable {
		log.Fatal("This node is not a swarm manager. The agent can only be run on a swarm manager.")
	}
	return cli
}

func CreateServiceSpec(
	imageName string,
	tag string,
	envVars map[string]string,
	Networks map[string]Network,
	NetworkList []string,
	secrets []string,
	volumes map[string]string,
	ports map[string]string,
	cli *client.Client,
) swarm.ServiceSpec {

	if tag == "" {
		tag = "latest"
	}

	env := []string{}
	for k, v := range envVars {
		env = append(env, k+"="+v)
	}

	// Network can be either defined as a list or map
	// handle both cases.
	networkConfigs := []swarm.NetworkAttachmentConfig{}

	var alias string = ""
	if len(NetworkList) > 0 {
		alias = LastPartAfterSlash(imageName)

		for _, network := range NetworkList {
			networkConfigs = append(networkConfigs, swarm.NetworkAttachmentConfig{
				Target:  network,
				Aliases: []string{alias},
			})
		}
	} else if len(Networks) > 0 {

		for key, value := range Networks {
			networkConfigs = append(networkConfigs, swarm.NetworkAttachmentConfig{
				Target:  key,
				Aliases: value.Aliases,
			})
		}
	} else {
		log.Error("No network config defined")
	}

	secretRefs := []*swarm.SecretReference{}
	for _, secret := range secrets {
		id, err := GetSecretIDByName(cli, secret)
		if err != nil {
			log.Fatalf("Secret does not exist, %s", err)
		}

		secretRefs = append(secretRefs, &swarm.SecretReference{
			SecretName: secret,
			SecretID:   id,
			File: &swarm.SecretReferenceFileTarget{
				Name: fmt.Sprintf("/run/secrets/%s", secret), // This should be just the filename, not the full path
				UID:  "0",
				GID:  "0",
				Mode: 0444,
			},
		})
	}

	mounts := []mount.Mount{}
	for src, target := range volumes {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeVolume,
			Source: src,
			Target: target,
		})
	}

	portConfigs := []swarm.PortConfig{}
	for published, target := range ports {
		publishedPort, err := strconv.ParseUint(published, 10, 16)
		if err != nil {
			log.Fatalf("Invalid published port %s: %v", published, err)
		}
		targetPort, err := strconv.ParseUint(target, 10, 16)
		if err != nil {
			log.Fatalf("Invalid target port %s: %v", target, err)
		}

		portConfigs = append(portConfigs, swarm.PortConfig{
			Protocol:      swarm.PortConfigProtocolTCP,
			PublishedPort: uint32(publishedPort),
			TargetPort:    uint32(targetPort),
		})
	}

	return swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name: imageName,
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image:   imageName + ":" + tag,
				Env:     env,
				Secrets: secretRefs,
				Mounts:  mounts,
			},
			Networks: networkConfigs,
		},
		EndpointSpec: &swarm.EndpointSpec{
			Mode:  swarm.ResolutionModeVIP,
			Ports: portConfigs,
		},
	}
}

func GetSecretIDByName(cli *client.Client, secretName string) (string, error) {
	secrets, err := cli.SecretList(context.Background(), types.SecretListOptions{})
	if err != nil {
		return "", err
	}

	for _, secret := range secrets {
		if secret.Spec.Name == secretName {
			return secret.ID, nil
		}
	}

	return "", fmt.Errorf("secret not found: %s", secretName)
}

func CreateDockerService(cli *client.Client, spec swarm.ServiceSpec) types.ServiceCreateResponse {
	serviceSpecJSON, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling service spec to JSON: %v", err)
	}

	log.Println("---------------------------")
	log.Println(string(serviceSpecJSON))
	log.Println("---------------------------")

	// Create the service
	response, err := cli.ServiceCreate(context.Background(), spec, types.ServiceCreateOptions{})
	if err != nil {
		log.Fatalf("Error creating service: %v", err)
	}

	// Print the service ID
	log.WithFields(logrus.Fields{
		"responseId": response.ID,
	}).Info("Service created")

	return response
}
