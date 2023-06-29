package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

func createAcceptedDataRequest(validationResponse *pb.ValidationResponse, w http.ResponseWriter, authorizedProviders *map[string]string) {
	logger.Debug("Entering createAcceptedDataRequest")

	result := &pb.AcceptedDataRequest{}

	result.Auth = &pb.Auth{}
	result.User = &pb.User{}

	result.Auth = validationResponse.Auth
	result.User = validationResponse.User

	result.AuthorizedProviders = make(map[string]string)
	// Careful, not a deepcopy!
	result.AuthorizedProviders = *authorizedProviders
	result.ResultChannel = "tmp"

	for key, _ := range validationResponse.ValidDataproviders {
		var agentData lib.AgentDetails
		json, err := etcd.GetAndUnmarshalJSON(etcdClient, fmt.Sprintf("/agents/%s", key), &agentData)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		} else if json == nil {
			// invalidProviders = append(invalidProviders, key)
			continue
		}
		result.AuthorizedProviders[key] = agentData.Dns
	}
	if len(result.AuthorizedProviders) == 0 {

		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Request was processed, but no agreements or available dataproviders have been found"))
		return
	}

	jsonResponse, err := json.Marshal(result)
	if err != nil {
		logger.Sugar().Errorf("Error marshalling result, %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Debug("Response:")
	logger.Debug(string(jsonResponse))
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
