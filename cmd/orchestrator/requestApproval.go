package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Jorrit05/micro-recomposer/pkg/lib"
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
			log.Printf("%s: Error unmarshalling body: %v", serviceName, err)
			http.Error(w, "Error unmarshalling request body", http.StatusBadRequest)
			return
		}
		switch msgType.Type {
		case "requestApproval":

			var requestApproval lib.RequestApproval
			err = json.Unmarshal(body, &requestApproval)
			if err != nil {
				log.Printf("%s: Error unmarshalling body into RequestApproval: %v", serviceName, err)
				http.Error(w, "Error unmarshalling request body", http.StatusBadRequest)
				return
			}

			jsonRA, err := json.Marshal(requestApproval)
			if err != nil {
				log.Printf("%s: Error unmarshalling body into RequestApproval: %v", serviceName, err)
				http.Error(w, "Error unmarshalling request body", http.StatusBadRequest)
				return
			}

			url := policyEnforcerEndpoint + "/validate"
			approval, err := lib.PostRequest(url, string(jsonRA))
			if err != nil {
				log.Printf("%s: Error unmarshalling body into RequestApproval: %v", serviceName, err)
				http.Error(w, "Error unmarshalling request body", http.StatusBadRequest)
				return
			}

			fmt.Println(string(approval))

			err = checkRequestApproval(&requestApproval)
			if err != nil {
				log.Printf("%s: checkRequestApproval: %v", serviceName, err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
			return

		default:
			// Respond with a 405 'Method Not Allowed' HTTP response if the method isn't supported
			http.Error(w, "Invalid msg", http.StatusNotFound)
			return
		}
	}
}
