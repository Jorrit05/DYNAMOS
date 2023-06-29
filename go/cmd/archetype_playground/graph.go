package main

import (
	"fmt"
	"os"
)

func generateGraphViz(requestType RequestType, services []MicroserviceMetada, filename string) error {
	// Create a map of services for easy access
	serviceMap := make(map[string]MicroserviceMetada)
	for _, service := range services {
		serviceMap[service.Name] = service
	}

	// Define sets for required and optional services
	requiredServices := make(map[string]bool)
	for _, serviceName := range requestType.RequiredServices {
		requiredServices[serviceName] = true
	}

	optionalServices := make(map[string]bool)
	for _, serviceName := range requestType.OptionalServices {
		optionalServices[serviceName] = true
	}

	// Start building the graph description
	graphDescription := "digraph G {\n"

	// Loop over all services and add their dependencies to the graph
	for _, service := range services {
		for _, output := range service.AllowedOutputs {
			// Determine the color and label of the edge
			// color := "black"
			label := ""
			if optionalServices[output] {
				// color := "gray"
				label = " [color=gray, fontcolor=gray, label=\"optional\"]"
			}

			// Add the edge to the graph
			graphDescription += fmt.Sprintf("  %s -> %s%s ;\n", service.Name, output, label)
		}

		// Determine the color of the node
		color := "black"
		style := "solid"
		if optionalServices[service.Name] {
			color = "gray"
			style = "dashed"
		}

		// Add the node to the graph
		graphDescription += fmt.Sprintf("  %s [color=%s, style=%s];\n", service.Name, color, style)
	}

	// End the graph description
	graphDescription += "}\n"

	// Write the graph description to a file
	err := os.WriteFile(filename, []byte(graphDescription), 0644)
	if err != nil {
		return err
	}

	return nil
}
