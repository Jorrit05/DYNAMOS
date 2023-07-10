package main

import (
	"encoding/json"
	"net/http"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

func createAcceptedDataRequest(validationResponse *pb.ValidationResponse, w http.ResponseWriter, userTargets map[string]string) {
	logger.Debug("Entering createAcceptedDataRequest")

	result := &pb.AcceptedDataRequest{}

	result.Auth = &pb.Auth{}
	result.User = &pb.User{}

	result.Auth = validationResponse.Auth
	result.User = validationResponse.User

	result.AuthorizedProviders = make(map[string]string)
	result.AuthorizedProviders = userTargets
	result.ResultChannel = "tmp"

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
