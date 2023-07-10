package mschain

import (
	"fmt"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
)

var (
	logger = lib.InitLogger(logLevel)
)

func GenerateChain(services []MicroserviceMetadata) ([]MicroserviceMetadata, error) {

	// Build a map of nodes for each service
	nodes := make(map[string]*Node)
	for i := range services {
		// node[queryService] = QueryService Chaindata
		nodes[services[i].Name] = &Node{Service: &services[i]}
	}

	// Add edges between nodes based on allowed outputs
	for _, service := range services {
		node := nodes[service.Name]
		for _, output := range service.AllowedOutputs {
			// logger.Debug("OUTPUT:" + output)
			if nodes[output] != nil {
				node.OutEdges = append(node.OutEdges, nodes[output])
			}
		}
	}

	// Perform a topological sort to determine the order of services
	order, err := topologicalSort(nodes)
	if err != nil {
		logger.Sugar().Errorf("Error in topological sort: %v", err)
		return nil, err
	}

	filteredOrder := []MicroserviceMetadata{}
	for _, node := range order {
		filteredOrder = append(filteredOrder, *node.Service)
	}

	return filteredOrder, nil
}

func topologicalSort(nodes map[string]*Node) ([]*Node, error) {
	order := []*Node{}
	visited := make(map[string]bool)
	temp := make(map[string]bool)

	var visit func(node *Node) error
	visit = func(node *Node) error {
		// Return an error if we've already visited this node in the current path (i.e., there's a cycle)
		if temp[node.Service.Name] {
			logger.Sugar().Error("cycle detected")
			return fmt.Errorf("cycle detected")
		}

		// If we've already visited this node in a previous path, there's no need to visit it again
		if !visited[node.Service.Name] {
			temp[node.Service.Name] = true

			for _, output := range node.OutEdges {
				err := visit(output)
				if err != nil {
					return err
				}
			}

			visited[node.Service.Name] = true
			temp[node.Service.Name] = false

			// Add the node to the order
			order = append([]*Node{node}, order...)
		}

		return nil
	}

	for _, node := range nodes {
		err := visit(node)
		if err != nil {
			return nil, err
		}
	}

	return order, nil
}

func splitServicesByArchetype(orderedServices []MicroserviceMetadata, computeProvider string) SplitServices {
	var splitServices SplitServices

	for _, service := range orderedServices {
		if service.Label == "DataProvider" || computeProvider == "DataProvider" {
			splitServices.DataProviderServices = append(splitServices.DataProviderServices, service)
		} else if service.Label == "ComputeProvider" {
			splitServices.ComputeProviderServices = append(splitServices.ComputeProviderServices, service)
		}
	}

	return splitServices
}

// Helper function to check if a slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
