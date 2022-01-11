package main

type Config struct {
    ListeningPort int             `json:"listeningPort"`
    Sites         map[string]Site `json:"sites"` // [api key] => hugo site
}
