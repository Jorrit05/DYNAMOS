package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	logger = lib.InitLogger(logLevel)
)

type Auth struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type SqlDataRequest struct {
	Type                string            `json:"type"`
	Query               string            `json:"query"`
	Algorithm           string            `json:"algorithm"`
	User                map[string]string `json:"user"`
	Auth                Auth              `json:"auth"`
	Providers           []string          `json:"providers"`
	AuthorizedProviders map[string]string `json:"authorizedProviders"`
	Options             Options           `json:"options"`
}

type User struct {
	Id       string `json:"ID"`
	UserName string `json:"userName"`
	// Other fields...
}

type RequestApproval struct {
	Type          string   `json:"type"`
	User          User     `json:"user"`
	DataProviders []string `json:"dataProviders"`
	// DataRequest   DataRequest `json:"dataRequest"`
	DataRequest json.RawMessage `json:"data_request"`
}

type DataRequestOptions struct {
	Options map[string]bool `json:"options"`
	// Algorithm string          `json:"algorithm"`
	// Query     string          `json:"query"`
}

type Options struct {
	Aggregate bool `json:"aggregate"`
	Graph     bool `json:"graph"`
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
	Name             string            `json:"name"`
	RequiredServices []string          `json:"requiredServices"`
	OptionalServices map[string]string `json:"optionalServices"`
}

type Archetypes struct {
	Archetypes []Archetype `json:"archetypes"`
}

type Archetype struct {
	Name            string `json:"name"`
	ComputeProvider string `json:"computeProvider"`
	ResultRecipient string `json:"resultRecipient"`
	Weight          int    `json:"weight"`
}

type MicroserviceMetadata struct {
	Name           string   `json:"name"`
	Label          string   `json:"label"`
	AllowedOutputs []string `json:"allowedOutputs"`
}

type OptionalServices struct {
	DataSteward string              `json:"data_steward"`
	Types       map[string][]string `json:"types"`
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
	trimmedPath := strings.TrimPrefix(req.URL.Path, etcdRoot) //fmt.Sprintf("%s/", etcdRoot))
	fmt.Println("trimmedPath: " + trimmedPath)
	fmt.Println("req.URL.Path: " + req.URL.Path)
	var jsonData []byte
	var err error
	var target *T
	switch trimmedPath {
	case "":
		fallthrough
	case "/":
		logger.Info("Start GetPrefixListEtcd")
		targetList, err := etcd.GetPrefixListEtcd(etcdClient, etcdRoot, &target)

		if err != nil {
			logger.Sugar().Infof("Error in requesting config: %s", err)
			http.Error(w, "Error in requesting config", http.StatusInternalServerError)
			return
		}
		jsonData, err = json.Marshal(&targetList)
		if err != nil {
			logger.Sugar().Fatalw("Failed to convert map to JSON: %v", err)
		}

	default:
		key := fmt.Sprintf("%s%s", etcdRoot, trimmedPath)
		fmt.Println(key)
		jsonData, err = etcd.GetAndUnmarshalJSON(etcdClient, key, &target)

		if err != nil {
			logger.Sugar().Infof("Unknown path: %s", trimmedPath)
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
		logger.Sugar().Infof("handler: Error reading body: %v", err)
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
		logger.Sugar().Infof("Error in saving the  new archetype: %s", err)
		http.Error(w, "Error in saving the  new archetype", http.StatusInternalServerError)
		return
	}

	logger.Sugar().Infof("Added %s", key)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func PostRequest(url string, body string, extra_headers map[string]string) ([]byte, error) {
	reqBody := bytes.NewBufferString(body)
	req, err := http.NewRequest(http.MethodPost, url, reqBody)
	if err != nil {
		logger.Sugar().Infof("Failed to create request: %v", err)
		return []byte(""), err
	}

	headers := map[string]string{
		"Content-Type": "application/json",
		// add other headers as required
	}

	maps.Copy(headers, extra_headers)

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Sugar().Infof("Failed to make request: %v", err)
		return []byte(""), err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Sugar().Infof("Failed to read response body: %v", err)
		return []byte(""), err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		err = fmt.Errorf(fmt.Sprintf("Bad response from server: %s", resp.Status))
		logger.Sugar().Infof("%v", err)
		return []byte(""), err
	}

	return respBody, nil
}
