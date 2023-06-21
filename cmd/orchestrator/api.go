package main

import (
	"net/http"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func archetypesHandler(etcdClient *clientv3.Client, root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Entering archetypesHandler")
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

func agreementsHandler(etcdClient *clientv3.Client, etcdRoot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Call your handler for GET
			lib.GenericGetHandler[lib.Agreement](w, r, etcdClient, etcdRoot)
		case http.MethodPut:
			// Call your handler for PUT
			agreement := &lib.Agreement{}
			lib.GenericPutToEtcd[lib.Agreement](w, r, etcdClient, etcdRoot, agreement)
		default:
			// Respond with a 405 'Method Not Allowed' HTTP response if the method isn't supported
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
