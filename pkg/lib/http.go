package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func GenericGetHandler[T any](w http.ResponseWriter, req *http.Request, etcdClient *clientv3.Client, root string) {
	path := strings.TrimPrefix(req.URL.Path, fmt.Sprintf("%s/", root))

	var jsonData []byte
	var err error
	var target *T
	switch path {
	case "":
		targetList, err := GetPrefixListEtcd(etcdClient, root, &target)

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
		key := fmt.Sprintf("%s/%s", root, path)
		jsonData, err = GetAndUnmarshalJSON(etcdClient, key, &target)

		if err != nil {
			log.Printf("Unknown path: %s", path)
			http.Error(w, "Unknown request", http.StatusNotFound)
			return
		}

	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonData))
}
