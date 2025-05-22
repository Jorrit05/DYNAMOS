package mschain

type RequestType struct {
	Type             string            `json:"type"`
	RequiredServices []string          `json:"requiredServices"`
	OptionalServices map[string]string `json:"optionalServices"`
}

type MicroserviceMetadata struct {
	Name              string   `json:"name"`
	Label             string   `json:"label"`
	AllowedOutputs    []string `json:"allowedOutputs"`
	InvalidArchetypes []string `json:"invalidArchetypes"`
}

type Node struct {
	Service  *MicroserviceMetadata
	OutEdges []*Node
}

type SplitServices struct {
	DataProviderServices    []MicroserviceMetadata
	ComputeProviderServices []MicroserviceMetadata
}
