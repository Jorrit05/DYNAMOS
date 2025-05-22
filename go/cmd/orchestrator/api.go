package main

import (
	"net/http"

	"github.com/Jorrit05/DYNAMOS/pkg/api"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func archetypesHandler(etcdClient *clientv3.Client, root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Entering archetypesHandler")
		switch r.Method {
		case http.MethodGet:
			// Call your handler for GET
			api.GenericGetHandler[api.Archetype](w, r, etcdClient, "/archetypes")
		case http.MethodPut:
			// Call your handler for PUT
			archetype := &api.Archetype{}
			api.GenericPutToEtcd[api.Archetype](w, r, etcdClient, "/archetypes", archetype)
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
			api.GenericGetHandler[api.RequestType](w, r, etcdClient, etcdRoot)
		case http.MethodPut:
			// Call your handler for PUT
			requestType := &api.RequestType{}
			api.GenericPutToEtcd[api.RequestType](w, r, etcdClient, etcdRoot, requestType)
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
			api.GenericGetHandler[api.MicroserviceMetadata](w, r, etcdClient, etcdRoot)
		case http.MethodPut:
			// Call your handler for PUT
			msMetadata := &api.MicroserviceMetadata{}
			api.GenericPutToEtcd[api.MicroserviceMetadata](w, r, etcdClient, etcdRoot, msMetadata)
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
			api.GenericGetHandler[api.Agreement](w, r, etcdClient, etcdRoot)
		case http.MethodPut:
			// Call your handler for PUT
			agreement := &api.Agreement{}
			api.GenericPutToEtcd[api.Agreement](w, r, etcdClient, etcdRoot+"/agreements", agreement)
			go checkJobs(agreement)
		default:
			// Respond with a 405 'Method Not Allowed' HTTP response if the method isn't supported
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func updateEtc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Your original updateEtc code here.
		go registerPolicyEnforcerConfiguration()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Updated all config"))
	}
}
