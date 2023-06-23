package api

import (
	"io"
	"net/http"
)

func GetRequestBody(w http.ResponseWriter, r *http.Request, serviceName string) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	r.Body.Close()

	if err != nil {
		logger.Sugar().Infof("%s: Error reading body: %v", serviceName, err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return nil, err
	}

	return body, nil
}

func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request
		logger.Sugar().Infow("Received request", "method", r.Method, "path", r.URL.Path)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
