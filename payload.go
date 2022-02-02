package midas

type Payload interface {
	Event() string
	Metadata() map[string]interface{}
	Entry() map[string]interface{}
	Raw() interface{}
}
