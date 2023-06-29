package main

import (
	"context"
	"fmt"

	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

func startCompositionRequest(validationResponse *pb.ValidationResponse, authorizedProviders *map[string]string) error {
	logger.Debug("Entering startCompositionRequest")

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

	msChain, err := generateChain([]string{}, msMetadata)
	if err != nil {
		return err
	}

	//	{
	//	    "archetypeId": "computeToData",
	//	    "requestType": "sqlDataRequest",
	//	    "microservices" : ["queryService", "algorithmService"],
	//	    "user": {
	//	        "ID": "<GUID>",
	//	        "userName": "jorrit.stutterheim@cloudnation.nl"
	//	    },
	//	    "dataProvider" : ""
	//	}
	//TODO: aanpassen
	compositionRequest := &pb.CompositionRequest{}
	compositionRequest.User = &pb.User{}
	compositionRequest.User = validationResponse.User
	compositionRequest.DataProvider = ""
	compositionRequest.ArchetypeId = chooseArchetype(validationResponse)
	compositionRequest.RequestType = "sqlDataRequest"
	compositionRequest.Target = "UVA-in"
	for _, chain := range msChain {
		compositionRequest.Microservices = append(compositionRequest.Microservices, chain.Name)
	}

	c.SendCompositionRequest(context.Background(), compositionRequest)

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
