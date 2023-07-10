package mschain

type RequestType struct {
	Type             string   `json:"type"`
	RequiredServices []string `json:"requiredServices"`
	OptionalServices []string `json:"optionalServices"`
}

type MicroserviceMetadata struct {
	Name           string   `json:"name"`
	Label          string   `json:"label"`
	AllowedOutputs []string `json:"AllowedOutputs"`
}

type Node struct {
	Service  *MicroserviceMetadata
	OutEdges []*Node
}
type SplitServices struct {
	DataProviderServices    []MicroserviceMetadata
	ComputeProviderServices []MicroserviceMetadata
}
