package main

import (
	"context"
	"encoding/json"
	"net/http"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"go.opencensus.io/trace"
)

func createAcceptedDataRequest(ctx context.Context, validationResponse *pb.ValidationResponse, w http.ResponseWriter, userTargets map[string]string) context.Context {
	logger.Debug("Entering createAcceptedDataRequest")
	ctx, span := trace.StartSpan(ctx, "createAcceptedDataRequest")
	defer span.End()

	result := &pb.AcceptedDataRequest{}

	result.Auth = &pb.Auth{}
	result.User = &pb.User{}

	result.Auth = validationResponse.Auth
	result.User = validationResponse.User

	result.AuthorizedProviders = make(map[string]string)
	result.AuthorizedProviders = userTargets

	jsonResponse, err := json.Marshal(result)
	if err != nil {
		logger.Sugar().Errorf("Error marshalling result, %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return ctx
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
	return ctx
}
