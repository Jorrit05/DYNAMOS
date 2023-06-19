package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type User struct {
	ID       string `json:"ID"`
	UserName string `json:"userName"`
	// Other fields...
}

type RequestApproval struct {
	Type          string   `json:"type"`
	User          User     `json:"user"`
	DataProviders []string `json:"dataProviders"`
	SyncServices  bool     `json:"syncServices"`
}

type Relation struct {
	ID                      string   `json:"ID"`
	RequestTypes            []string `json:"requestTypes"`
	DataSets                []string `json:"dataSets"`
	AllowedArchetypes       []string `json:"allowedArchetypes"`
	AllowedComputeProviders []string `json:"allowedComputeProviders"`
}

type Agreement struct {
	Name             string              `json:"name"`
	Relations        map[string]Relation `json:"relations"`
	ComputeProviders []string            `json:"computeProviders"`
	Archetypes       []string            `json:"archetypes"`
}

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

type Auth struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type ValidationResponse struct {
	Type                    string   `json:"type"`
	RequestType             string   `json:"requestType"`
	ValidDataProviders      []string `json:"validDataProviders"`
	InvalidDataProviders    []string `json:"invalidDataProviders"`
	Auth                    Auth     `json:"auth"`
	AllowedArcheTypes       []string `json:"allowedArcheTypes"`
	AllowedComputeProviders []string `json:"allowedComputeProviders"`
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

func (a Agreement) GetName() string {
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
			logger.Sugar().Infow("Error in requesting config: %s", err)
			http.Error(w, "Error in requesting config", http.StatusInternalServerError)
			return
		}
		jsonData, err = json.Marshal(&targetList)
		if err != nil {
			logger.Sugar().Fatalw("Failed to convert map to JSON: %v", err)
		}

	default:
		key := fmt.Sprintf("%s/%s", etcdRoot, trimmedPath)
		fmt.Println(key)
		jsonData, err = GetAndUnmarshalJSON(etcdClient, key, &target)

		if err != nil {
			logger.Sugar().Infow("Unknown path: %s", trimmedPath)
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
	//TODO:
	// Allow longer ETCD paths. Now /policyEnforcer/agreements/VU, will be put at /policyEnforcer/VU. Probably insert trimmedPath.
	// First write unit tests though.

	body, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		logger.Sugar().Infow("handler: Error reading body: %v", err)
		http.Error(w, "handler: Error reading request body", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &target)
	if err != nil {
		logger.Sugar().Errorw("failed to marshal struct: %v", err)
		http.Error(w, "Failed parsing body", http.StatusBadRequest)
		return
	}

	name := target.GetName()
	if name == "" {
		logger.Sugar().Errorw("Body does not have a name.: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// This seems like double work, but ensures only values of the struct are added.
	// (even if they are empty)
	jsonRep, err := json.Marshal(target)
	if err != nil {
		logger.Sugar().Errorw("failed to marshal struct: %v", err)
		http.Error(w, "Failed parsing body", http.StatusBadRequest)
		return
	}

	key := fmt.Sprintf("%s/%s", etcdRoot, name)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Save the JSON representation to the etcd key-value store
	_, err = etcdClient.Put(ctx, key, string(jsonRep))

	if err != nil {
		logger.Sugar().Infow("Error in saving the  new archetype: %s", err)
		http.Error(w, "Error in saving the  new archetype", http.StatusInternalServerError)
		return
	}

	logger.Sugar().Infow("Added %s", key)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func PostRequest(url string, body string) ([]byte, error) {
	reqBody := bytes.NewBufferString(body)
	req, err := http.NewRequest(http.MethodPost, url, reqBody)
	if err != nil {
		logger.Sugar().Infow("Failed to create request: %v", err)
		return []byte(""), err
	}

	headers := map[string]string{
		"Content-Type": "application/json",
		// add other headers as required
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Sugar().Infow("Failed to make request: %v", err)
		return []byte(""), err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Sugar().Infow("Failed to read response body: %v", err)
		return []byte(""), err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		err = fmt.Errorf(fmt.Sprintf("Bad response from server: %s", resp.Status))
		logger.Sugar().Infow("%v", err)
		return []byte(""), err
	}

	return respBody, nil
}
