package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Jorrit05/micro-recomposer/pkg/lib"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func archetypesHandler(etcdClient *clientv3.Client, root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Call your handler for GET
			lib.GenericGetHandler[Archetype](w, r, etcdClient, "/archetypes")
		case http.MethodPut:
			// Call your handler for PUT
			archetype := &Archetype{}
			GenericPutToEtcd[Archetype](w, r, etcdClient, "/archetypes", archetype)
		default:
			// Respond with a 405 'Method Not Allowed' HTTP response if the method isn't supported
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// Only works if struct has a name field
func GenericPutToEtcd[T any](w http.ResponseWriter, req *http.Request, etcdClient *clientv3.Client, root string, target Named) {

	body, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		log.Printf("handler: Error reading body: %v", err)
		http.Error(w, "handler: Error reading request body", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &target)
	if err != nil {
		log.Errorf("failed to marshal struct: %v", err)
		http.Error(w, "Failed parsing body", http.StatusBadRequest)
		return
	}

	name := target.GetName()
	if name == "" {
		log.Errorf("Body does not have a name.: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// This seems like double work, but ensures only values of the struct are added.
	// (even if they are empty)
	jsonRep, err := json.Marshal(target)
	if err != nil {
		log.Errorf("failed to marshal struct: %v", err)
		http.Error(w, "Failed parsing body", http.StatusBadRequest)
		return
	}

	key := fmt.Sprintf("%s/%s", root, name)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Save the JSON representation to the etcd key-value store
	_, err = etcdClient.Put(ctx, key, string(jsonRep))

	if err != nil {
		log.Printf("Error in saving the  new archetype: %s", err)
		http.Error(w, "Error in saving the  new archetype", http.StatusInternalServerError)
		return
	}

	log.Infof("Added %s", key)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
