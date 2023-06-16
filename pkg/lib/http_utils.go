package lib

import (
	"io"
	"net/http"
)

func GetRequestBody(w http.ResponseWriter, r *http.Request, serviceName string) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	r.Body.Close()

	if err != nil {
		log.Printf("%s: Error reading body: %v", serviceName, err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return nil, err
	}

	return body, nil
}
