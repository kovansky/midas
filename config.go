package strapi2hugo

type Config struct {
	ListeningPort int             `json:"listeningPort"`
	Sites         map[string]Site `json:"sites"` // [api key] => site
}
