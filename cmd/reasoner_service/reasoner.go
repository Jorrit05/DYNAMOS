package main

import (
	"encoding/json"
	"fmt"
	"io"

	"net/http"
	"os"

	"github.com/Jorrit05/micro-recomposer/pkg/lib"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	serviceName                   = "reasoner_service"
	log, logFile                  = lib.InitLogger(serviceName)
	etcdClient   *clientv3.Client = lib.GetEtcdClient()
)

type Updatable interface {
	GetName() string
}

type updateRequest struct {
	Type          string        `json:"type"`
	RequestorName string        `json:"requestor"`
	Archetype     lib.ArcheType `json:"archetype"`
	RequestorData lib.Requestor `json:"requestor_data"`
}

func main() {
	defer logFile.Close()
	defer etcdClient.Close()
	defer lib.HandlePanicAndFlushLogs(log, logFile)

	registerReasonerConfiguration()

	mux := http.NewServeMux()
	mux.HandleFunc("/update", updateHandler)
	mux.HandleFunc("/archetype_config", getHandler)
	mux.HandleFunc("/requestor_config", getHandler)
	log.Info("Starting http server on 8081/30011")
	go func() {
		if err := http.ListenAndServe(":8081", mux); err != nil {
			log.Fatalf("Error starting HTTP server: %s", err)
		}
	}()

	select {}

}

func registerReasonerConfiguration() {

	archetypesJSON, err := os.ReadFile("/var/log/stack-files/config/archetype_config.json")
	if err != nil {
		log.Fatalf("Failed to read the JSON archetype config file: %v", err)
	}

	lib.RegisterJSONArray[lib.ArcheType](archetypesJSON, &lib.ArcheTypes{}, etcdClient, "/reasoner/archetype_config")

	requestorConfigJSON, err := os.ReadFile("/var/log/stack-files/config/requestor_config.json")
	if err != nil {
		log.Fatalf("Failed to read the JSON requestor config file: %v", err)
	}

	lib.RegisterJSONArray[lib.Requestor](requestorConfigJSON, &lib.RequestorConfig{}, etcdClient, "/reasoner/requestor_config")
}

func getHandler(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	key := fmt.Sprintf("/reasoner%s/", path)
	var jsonData []byte

	switch path {
	case "/archetype_config":
		archetypeConf, err := lib.GetAndUnmarshalJSONMap[lib.ArcheType](etcdClient, key)

		if err != nil {
			log.Printf("Error in requesting config: %s", err)
			http.Error(w, "Error in requesting config", http.StatusInternalServerError)
			return
		}
		jsonData, err = json.Marshal(archetypeConf)
		if err != nil {
			log.Fatalf("Failed to convert map to JSON: %v", err)
		}

	case "/requestor_config":

		requestorConf, err := lib.GetAndUnmarshalJSONMap[lib.Requestor](etcdClient, key)
		if err != nil {
			log.Printf("Error in requesting config: %s", err)
			http.Error(w, "Error in requesting config", http.StatusInternalServerError)
			return
		}

		jsonData, err = json.Marshal(requestorConf)
		if err != nil {
			log.Fatalf("Failed to convert map to JSON: %v", err)
		}

	default:
		log.Printf("Unknown path: %s", path)
		http.Error(w, "Unknown request", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonData))
}

func updateHandler(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("handler: Error reading body: %v", err)
		http.Error(w, "handler: Error reading request body", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	// Unmarshal request body
	var updateReq updateRequest
	err = json.Unmarshal(body, &updateReq)
	if err != nil {
		http.Error(w, "Failed to unmarshal request body", http.StatusBadRequest)
		return
	}

	// Process the request based on the "type" field
	switch updateReq.Type {
	case "archeTypeUpdate":
		updateArchetypeHandler(updateReq, w, req)
	case "requestorUpdate":
		updateRequestorHandler(updateReq, w, req)

	default:
		log.Printf("Unknown message type: %s", updateReq.Type)
		http.Error(w, "Unknown request", http.StatusNotFound)
		return
	}
}

func updateArchetypeHandler(updateReq updateRequest, w http.ResponseWriter, req *http.Request) {
	// Load archetypes from etcd
	archeTypeMap, err := lib.GetAndUnmarshalJSONMap[lib.ArcheType](etcdClient, "/reasoner/archetype_config/")
	if err != nil {
		http.Error(w, "Failed get data", http.StatusInternalServerError)
	}

	// Find and update the target requestor
	archeTypeMap[updateReq.Archetype.Name] = updateArchetype(archeTypeMap[updateReq.Archetype.Name], updateReq.Archetype)

	// Save updated requestors back to etcd
	err = lib.SaveStructToEtcd(etcdClient, "/reasoner/archetype_config/"+updateReq.Archetype.Name, archeTypeMap[updateReq.Archetype.Name])
	if err != nil {
		http.Error(w, "Failed to save updated ar to etcd", http.StatusInternalServerError)
		return
	}
	log.Info("Saved update with name %s as new config in '/reasoner/archetype_config/'.", updateReq.Archetype.Name)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Archetype updated successfully"))
}

func updateArchetype(old, new lib.ArcheType) lib.ArcheType {
	if new.RequestType != "" {
		old.RequestType = new.RequestType
	}

	if len(new.IoConfig.ServiceIO) > 0 {
		old.IoConfig.ServiceIO = new.IoConfig.ServiceIO
	}

	if len(new.IoConfig.ThirdParty) > 0 {
		old.IoConfig.ThirdParty = new.IoConfig.ThirdParty
	}

	if new.IoConfig.Finish != "" {
		old.IoConfig.Finish = new.IoConfig.Finish
	}

	if new.IoConfig.ThirdPartyName != "" {
		old.IoConfig.ThirdPartyName = new.IoConfig.ThirdPartyName
	}

	return old
}

func updateRequestorHandler(updateReq updateRequest, w http.ResponseWriter, req *http.Request) {
	// Load requestors from etcd
	requestorMap, err := lib.GetAndUnmarshalJSONMap[lib.Requestor](etcdClient, "/reasoner/requestor_config/")
	if err != nil {
		http.Error(w, "Failed to get data", http.StatusInternalServerError)
	}

	// Find and update the target requestor
	requestorMap[updateReq.RequestorData.Name] = updateRequestor(requestorMap[updateReq.RequestorData.Name], updateReq.RequestorData)

	// Save updated requestors back to etcd
	err = lib.SaveStructToEtcd(etcdClient, "/reasoner/requestor_config/"+updateReq.RequestorData.Name, requestorMap[updateReq.RequestorData.Name])
	if err != nil {
		http.Error(w, "Failed to save updated requestors to etcd", http.StatusInternalServerError)
		return
	}

	log.Info("Saved update with name %s as new config in '/reasoner/requestor_config/'.", updateReq.RequestorData.Name)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Requestor updated successfully"))
}

func updateRequestor(old, new lib.Requestor) lib.Requestor {
	if new.CurrentArchetype != "" {
		old.CurrentArchetype = new.CurrentArchetype
	}

	if len(new.AllowedPartners) > 0 {
		old.AllowedPartners = new.AllowedPartners
	}

	return old
}
