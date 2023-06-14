package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type RequestType struct {
	Name             string   `json:"name"`
	RequiredServices []string `json:"requiredServices"`
	OptionalServices []string `json:"optionalServices"`
}

type Archetype struct {
	Name            string `json:"name"`
	ComputeProvider string `json:"computeProvider"`
	ResultRecipient string `json:"resultRecipient"`
}

type MicroserviceMetadata struct {
	Name           string   `json:"name"`
	Label          string   `json:"label"`
	AllowedOutputs []string `json:"allowedOutputs"`
}

type Named interface {
	GetName() string
}

func (a Archetype) GetName() string {
	return a.Name
}

func (a RequestType) GetName() string {
	return a.Name
}

func (a MicroserviceMetadata) GetName() string {
	return a.Name
}

func GenericGetHandler[T any](w http.ResponseWriter, req *http.Request, etcdClient *clientv3.Client, etcdRoot string) {
	trimmedPath := strings.TrimPrefix(req.URL.Path, fmt.Sprintf("%s/", etcdRoot))
	fmt.Println("trimmedPath: " + trimmedPath)
	var jsonData []byte
	var err error
	var target *T
	switch trimmedPath {
	case "":
		targetList, err := GetPrefixListEtcd(etcdClient, etcdRoot, &target)

		if err != nil {
			log.Printf("Error in requesting config: %s", err)
			http.Error(w, "Error in requesting config", http.StatusInternalServerError)
			return
		}
		jsonData, err = json.Marshal(&targetList)
		if err != nil {
			log.Fatalf("Failed to convert map to JSON: %v", err)
		}

	default:
		key := fmt.Sprintf("%s/%s", etcdRoot, trimmedPath)
		jsonData, err = GetAndUnmarshalJSON(etcdClient, key, &target)

		if err != nil {
			log.Printf("Unknown path: %s", trimmedPath)
			http.Error(w, "Unknown request", http.StatusNotFound)
			return
		}

	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonData))
}

// Updata a single JSON struct to Etcd. (not many validity checks)
// Only works if struct has a name field and target has implemented the Named interface.
func GenericPutToEtcd[T any](w http.ResponseWriter, req *http.Request, etcdClient *clientv3.Client, etcdRoot string, target Named) {

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

	key := fmt.Sprintf("%s/%s", etcdRoot, name)
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
