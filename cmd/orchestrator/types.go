package main

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

type Node struct {
	Service  *MicroserviceMetada
	OutEdges []*Node
}
type SplitServices struct {
	DataProviderServices    []MicroserviceMetada
	ComputeProviderServices []MicroserviceMetada
}
