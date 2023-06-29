package main

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

func monitorServices(serviceId string) {

	// List tasks for the service
	tasks, err := dockerClient.TaskList(context.Background(), types.TaskListOptions{Filters: filters.NewArgs(filters.Arg("service", serviceId))})
	if err != nil {
		panic(err)
	}

	// Iterate through the tasks for the service
	for _, task := range tasks {
		taskID := task.ID
		taskState := task.Status.State

		fmt.Printf("Task ID: %s, Task State: %s\n", taskID, taskState)

		// Print the environment variables for each task
		for _, envVar := range task.Spec.ContainerSpec.Env {
			fmt.Println(envVar)
		}
	}
}
