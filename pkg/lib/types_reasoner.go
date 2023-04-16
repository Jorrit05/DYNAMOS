package lib

var _ NameGetter = (*Requestor)(nil)

type Iterable interface {
	Len() int
	Get(index int) interface{}
}

type NameGetter interface {
	GetName() string
}

type Requestor struct {
	Name             string   `json:"name"`
	CurrentArchetype string   `json:"current_archetype"`
	AllowedPartners  []string `json:"allowed_partners"`
}

type RequestorConfig struct {
	Contents []Requestor `json:"requestor_config"`
}

func (c *RequestorConfig) Len() int {
	return len(c.Contents)
}

func (c *RequestorConfig) Get(index int) interface{} {
	return &c.Contents[index]
}

func (a *Requestor) GetName() string {
	return a.Name
}
