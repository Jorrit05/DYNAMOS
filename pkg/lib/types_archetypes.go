package lib

var _ NameGetter = (*ArcheType)(nil)

type ArcheTypes struct {
	Contents []ArcheType `json:"archetypes"`
}

type ArcheType struct {
	Name        string   `json:"name"`
	RequestType string   `json:"request_type"`
	IoConfig    IoConfig `json:"io_config"`
}

type IoConfig struct {
	ServiceIO      map[string]string `json:"service_io"`
	Finish         string            `json:"finish"`
	ThirdPartyName string            `json:"third_party_name"`
	ThirdParty     map[string]string `json:"third_party"`
}

func (c *ArcheTypes) Len() int {
	return len(c.Contents)
}

//	func (c *ArcheTypes) Get(index int) interface{} {
//		return c.Contents[index]
//	}
func (c *ArcheTypes) Get(index int) interface{} {
	return &c.Contents[index]
}

func (a *ArcheType) GetName() string {
	return a.Name
}
