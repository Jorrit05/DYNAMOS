package main

import "net/http"

type TypeField struct {
	Type string `json:"type"`
}

func sqlDataRequestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Entering sqlDataRequestHandler")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("In correct switch statement"))

		// logger.Error("Unknown message type: " + typeField.Type)
		// http.Error(w, "Page not found", http.StatusNotFound)
	}
}
