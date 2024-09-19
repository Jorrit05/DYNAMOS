// This file contains the handlers for the requests that the API Gateway receives from the client
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/api"
	"go.opencensus.io/trace"
)

func sendData(endpoint string, jsonData []byte) (string, error) {
	// FIXME: Change to an actual token in the future?
	headers := map[string]string{
		"Authorization": "bearer 1234",
	}
	body, err := api.PostRequest(endpoint, string(jsonData), headers)
	if err != nil {
		return "", err
	}

	// Here we should send the request over the socket
	// For now we should append it to a list so that we gather all responses and send them in bulk
	logger.Sugar().Infof("Body: %v", body)
	return string(body), nil
}

func orchestratorHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func policyEnforcerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, testRequest, err := handleRequest(w, r, "policy-enforcer")
		if err != nil {
			return
		}

		switch testRequest.MethodName {
		case "":
		}
	}
}

func apiGatewayHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}

func handleRequest(w http.ResponseWriter, r *http.Request, requestType string) (context.Context, TestRequest, error) {
	ctxWithTimeout, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Start a new span with the context that has a timeout
	ctx, span := trace.StartSpan(ctxWithTimeout, requestType)
	defer span.End()

	body, err := api.GetRequestBody(w, r, serviceName)
	if err != nil {
		return ctx, TestRequest{}, err
	}

	var testRequest TestRequest
	if err := json.Unmarshal(body, &testRequest); err != nil {
		logger.Sugar().Errorf("Error unmMarshalling get testRequest for : %v", err)
		return ctx, TestRequest{}, err
	}

	return ctx, testRequest, nil
}
