package midas

type Config struct {
	Domain string          `json:"domain"`
	Addr   string          `json:"addr"`
	Sites  map[string]Site `json:"sites"` // [api key] => site
}
