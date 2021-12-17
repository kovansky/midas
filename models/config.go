package models

type Config struct {
	ListeningPort int                 `json:"listeningPort"`
	Sites         map[string]HugoSite `json:"sites"` // [api key] => hugo site
}
