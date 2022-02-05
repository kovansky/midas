package midas

type Config struct {
	Domain       string          `json:"domain"`
	Addr         string          `json:"addr"`
	RollbarToken string          `json:"rollbar_token"`
	Sites        map[string]Site `json:"sites"` // [api key] => site
}
