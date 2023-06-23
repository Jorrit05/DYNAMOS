package main

import (
	"fmt"

	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

type RequestType struct {
	Type             string   `json:"type"`
	RequiredServices []string `json:"requiredServices"`
	OptionalServices []string `json:"optionalServices"`
}

type MicroserviceMetada struct {
	Name           string   `json:"name"`
	Label          string   `json:"label"`
	AllowedOutputs []string `json:"AllowedOutputs"`
}

func startCompositionRequest(validationResponse *pb.ValidationResponse) error {
	logger.Debug("Entering startCompositionRequest")
	// archetype := chooseArchetype(validationResponse)
	var request RequestType
	_, err := etcd.GetAndUnmarshalJSON(etcdClient, fmt.Sprintf("/requestTypes/%s", validationResponse.RequestType), &request)
	if err != nil {
		return err
	}

	var msMetadata []MicroserviceMetada
	err = getMicroserviceMetadata(&msMetadata, &request)
	if err != nil {
		return err
	}
	logger.Sugar().Info(msMetadata[0])
	logger.Sugar().Info(msMetadata[1])
	logger.Sugar().Info(msMetadata[2])
	logger.Sugar().Info(msMetadata[3])
	return nil
}

func getMicroserviceMetadata(microserviceMetada *[]MicroserviceMetada, request *RequestType) error {

	for _, ms := range request.RequiredServices {
		var metadataObject MicroserviceMetada

		_, err := etcd.GetAndUnmarshalJSON(etcdClient, fmt.Sprintf("/microservices/%s/chainMetadata", ms), &metadataObject)
		if err != nil {
			return err
		}
		*microserviceMetada = append(*microserviceMetada, metadataObject)
	}

	for _, ms := range request.OptionalServices {
		var metadataObject MicroserviceMetada

		_, err := etcd.GetAndUnmarshalJSON(etcdClient, fmt.Sprintf("/microservices/%s/chainMetadata", ms), &metadataObject)
		if err != nil {
			return err
		}
		*microserviceMetada = append(*microserviceMetada, metadataObject)
	}

	return nil
}

func getRequestTypeMicroservices(requestType string) (RequestType, error) {
	var request RequestType
	_, err := etcd.GetAndUnmarshalJSON(etcdClient, fmt.Sprintf("/requestTypes/%s", requestType), &request)
	if err != nil {
		return RequestType{}, err
	}

	return request, nil
}

func chooseArchetype(validationResponse *pb.ValidationResponse) string {
	return "computeToData"
}
