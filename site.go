package midas

type Site struct {
	SiteName        string                   `json:"siteName"`
	RootDir         string                   `json:"rootDir"`
	Service         string                   `json:"service"`
	Registry        RegistrySettings         `json:"registry"`
	CollectionTypes map[string]ModelSettings `json:"collectionTypes"`
	SingleTypes     map[string]ModelSettings `json:"singleTypes"`
}

type ModelSettings struct {
	ArchetypePath string `json:"archetypePath,omitempty"`
	OutputDir     string `json:"outputDir,omitempty"`
}

type RegistrySettings struct {
	Type     string `json:"type"`
	Location string `json:"location"`
}

type SiteService interface {
	GetRegistryService() (RegistryService, error)
	BuildSite(useCache bool) error
	CreateEntry(payload Payload) (string, error)
	UpdateEntry(payload Payload) (string, error)
	DeleteEntry(payload Payload) (string, error)
}
