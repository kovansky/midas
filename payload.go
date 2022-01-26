package strapi2hugo

type Payload interface {
	Event() string
	Metadata() map[string]interface{}
	Entry() map[string]interface{}
	Raw() interface{}
}
