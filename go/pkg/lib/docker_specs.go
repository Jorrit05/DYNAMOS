package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
)

func GetDockerClient() *client.Client {
	// Create a new Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logger.Sugar().Fatalw("Error creating Docker client: %v", err)
	}

	info, err := cli.Info(context.Background())
	if err != nil {
		logger.Sugar().Fatalw("Error getting Docker info: %v", err)
	}

	fmt.Println(info)
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
	deployOpts Deploy,
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
		logger.Error("No network config defined")
	}

	secretRefs := []*swarm.SecretReference{}
	for _, secret := range secrets {
		id, err := GetSecretIDByName(cli, secret)
		if err != nil {
			logger.Sugar().Fatalw("Secret does not exist, %s", err)
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
			logger.Sugar().Fatalw("Invalid published port %s: %v", published, err)
		}
		targetPort, err := strconv.ParseUint(target, 10, 16)
		if err != nil {
			logger.Sugar().Fatalw("Invalid target port %s: %v", target, err)
		}

		portConfigs = append(portConfigs, swarm.PortConfig{
			Protocol:      swarm.PortConfigProtocolTCP,
			PublishedPort: uint32(publishedPort),
			TargetPort:    uint32(targetPort),
		})
	}

	// Here we assume we always want at least 1 replica.
	if deployOpts.Replicas == uint64(0) {
		deployOpts.Replicas = uint64(1)
	}

	mode := swarm.ServiceMode{
		Replicated: &swarm.ReplicatedService{
			Replicas: &deployOpts.Replicas,
		},
	}

	updateConfig := swarm.UpdateConfig{
		Parallelism: deployOpts.UpdateConfig.Parallelism,
		Delay:       deployOpts.UpdateConfig.Delay,
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
			Networks:  networkConfigs,
			Placement: &deployOpts.Placement,
			Resources: &deployOpts.Resources,
		},
		EndpointSpec: &swarm.EndpointSpec{
			Mode:  swarm.ResolutionModeVIP,
			Ports: portConfigs,
		},
		Mode:         mode,
		UpdateConfig: &updateConfig,
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
		logger.Sugar().Fatalw("Error marshaling service spec to JSON: %v", err)
	}

	fmt.Println("---------------------------")
	logger.Info(string(serviceSpecJSON))
	fmt.Println("---------------------------")

	// Create the service
	response, err := cli.ServiceCreate(context.Background(), spec, types.ServiceCreateOptions{})
	if err != nil {
		logger.Sugar().Fatalw("Error creating service: %v", err)
	}

	// Print the service ID
	logger.Sugar().Infof("ID: %s", response.ID)

	return response
}

func UpdateServiceReplicas(cli *client.Client, serviceID string, replicas uint64) error {
	// Get the service
	service, _, err := cli.ServiceInspectWithRaw(context.Background(), serviceID, types.ServiceInspectOptions{})
	if err != nil {
		logger.Sugar().Errorw("Error getting service: %v", err)
		return err
	}

	// Update the number of replicas
	service.Spec.Mode.Replicated.Replicas = &replicas

	// Prepare the service update options
	updateOpts := types.ServiceUpdateOptions{
		QueryRegistry: true,
	}

	// Update the service
	response, err := cli.ServiceUpdate(context.Background(), serviceID, service.Version, service.Spec, updateOpts)
	if err != nil {
		logger.Sugar().Errorw("Error updating service: %v", err)
		return err
	}

	logger.Sugar().Infof("responseWarnings: %s", strings.Join(response.Warnings, ","))

	return nil
}
