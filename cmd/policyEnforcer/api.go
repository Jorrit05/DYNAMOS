package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Jorrit05/micro-recomposer/pkg/lib"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func archetypesHandler(etcdClient *clientv3.Client, root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Call your handler for GET
			lib.GenericGetHandler[lib.Archetype](w, r, etcdClient, "/archetypes")
		case http.MethodPut:
			// Call your handler for PUT
			archetype := &lib.Archetype{}
			lib.GenericPutToEtcd[lib.Archetype](w, r, etcdClient, "/archetypes", archetype)
		default:
			// Respond with a 405 'Method Not Allowed' HTTP response if the method isn't supported
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func requestTypesHandler(etcdClient *clientv3.Client, etcdRoot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Call your handler for GET
			lib.GenericGetHandler[lib.RequestType](w, r, etcdClient, etcdRoot)
		case http.MethodPut:
			// Call your handler for PUT
			requestType := &lib.RequestType{}
			lib.GenericPutToEtcd[lib.RequestType](w, r, etcdClient, etcdRoot, requestType)
		default:
			// Respond with a 405 'Method Not Allowed' HTTP response if the method isn't supported
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func microserviceMetadataHandler(etcdClient *clientv3.Client, etcdRoot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Call your handler for GET
			lib.GenericGetHandler[lib.MicroserviceMetadata](w, r, etcdClient, etcdRoot)
		case http.MethodPut:
			// Call your handler for PUT
			msMetadata := &lib.MicroserviceMetadata{}
			lib.GenericPutToEtcd[lib.MicroserviceMetadata](w, r, etcdClient, etcdRoot, msMetadata)
		default:
			// Respond with a 405 'Method Not Allowed' HTTP response if the method isn't supported
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func agreementsHandler(etcdClient *clientv3.Client, root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Call your handler for GET
			lib.GenericGetHandler[lib.Agreement](w, r, etcdClient, "/agreements")
		case http.MethodPut:
			// Call your handler for PUT
			agreement := &lib.Agreement{}
			lib.GenericPutToEtcd[lib.Agreement](w, r, etcdClient, "/agreements", agreement)
		default:
			// Respond with a 405 'Method Not Allowed' HTTP response if the method isn't supported
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

type messageType struct {
	Type string `json:"type"`
}

func requestApprovalHandler(etcdClient *clientv3.Client, etcdRoot string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		r.Body.Close()

		if err != nil {
			log.Printf("handler: Error reading body: %v", err)
			http.Error(w, "handler: Error reading request body", http.StatusBadRequest)
			return
		}
		var msgType messageType
		err = json.Unmarshal(body, &msgType)
		if err != nil {
			log.Printf("handler: Error unmarshalling body: %v", err)
			http.Error(w, "handler: Error unmarshalling request body", http.StatusBadRequest)
			return
		}
		switch msgType.Type {
		case "requestApproval":
			// Call your handler for GET
			var requestApproval lib.RequestApproval
			err = json.Unmarshal(body, &requestApproval)
			if err != nil {
				log.Printf("handler: Error unmarshalling body into RequestApproval: %v", err)
				http.Error(w, "handler: Error unmarshalling request body", http.StatusBadRequest)
				return
			}

		default:
			// Respond with a 405 'Method Not Allowed' HTTP response if the method isn't supported
			http.Error(w, "Invalid type", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
