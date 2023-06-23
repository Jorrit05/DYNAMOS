package main

import (
	"encoding/json"
	"net/http"

	"github.com/Jorrit05/DYNAMOS/pkg/api"
)

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		body, err := api.GetRequestBody(w, r, serviceName)
		if err != nil {
			return
		}

		var sqlDataRequest api.SqlDataRequest
		err = json.Unmarshal(body, &sqlDataRequest)
		if err != nil {
			logger.Sugar().Errorf("Error unmarshalling SqlDataRequest: %v", err)
			return
		}

		if sqlDataRequest.Auth.AccessToken != "1234" {

			logger.Warn("Invalid token: " + sqlDataRequest.Auth.AccessToken)
			http.Error(w, "Invalid token", http.StatusForbidden)
			return
		}

		//TOD
		// Add auth value to request context for later use in handlers if needed
		// ctx := context.WithValue(r.Context(), "auth", auth)
		// r = r.WithContext(ctx)
		logger.Info("Serve next")
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
