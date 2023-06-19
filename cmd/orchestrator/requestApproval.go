package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
)

type messageType struct {
	Type string `json:"type"`
}

func requestApprovalHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := lib.GetRequestBody(w, r, serviceName)
		if err != nil {
			return
		}

		// Get the 'type' field of the message
		var msgType messageType
		err = json.Unmarshal(body, &msgType)
		if err != nil {
			logger.Sugar().Infow("%s: Error unmarshalling body: %v", serviceName, err)
			http.Error(w, "Error unmarshalling request body", http.StatusBadRequest)
			return
		}
		switch msgType.Type {
		case "requestApproval":

			// Convert to struct
			var requestApproval lib.RequestApproval
			err = json.Unmarshal(body, &requestApproval)
			if err != nil {
				logger.Sugar().Infow("%s: Error unmarshalling body into RequestApproval: %v", serviceName, err)
				http.Error(w, "Error unmarshalling request body", http.StatusBadRequest)
				return
			}

			handleRequestApproval(w, &requestApproval)
			return

		default:
			// Respond with a 405 'Method Not Allowed' HTTP response if the method isn't supported
			http.Error(w, "Invalid msg", http.StatusNotFound)
			return
		}
	}
}

func handleRequestApproval(w http.ResponseWriter, requestApproval *lib.RequestApproval) {

	// Convert back to JSON to pass on to the policy enforcer
	jsonRA, err := json.Marshal(requestApproval)
	if err != nil {
		logger.Sugar().Errorw("%s: Error unmarshalling body into RequestApproval: %v", serviceName, err)
		http.Error(w, "Error unmarshalling request body", http.StatusBadRequest)
		return
	}

	// Change this to Rabbit?
	// Get an approval.
	//   - I can have an agreement with each data steward
	//   - Get a result channel or endpoint
	//   - Return an access token
	//   - Start a composition request
	url := policyEnforcerEndpoint + "/validate"
	approval, err := lib.PostRequest(url, string(jsonRA))
	if err != nil {
		logger.Sugar().Errorw("%s: Error unmarshalling body into RequestApproval: %v", serviceName, err)
		http.Error(w, "Error unmarshalling request body", http.StatusBadRequest)
		return
	}

	fmt.Println(string(approval))
	var validationResponse lib.ValidationResponse
	fmt.Println(validationResponse)

	// Start orchestration request
	// Use archetypeplayground
	// Compose message back to user.

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
