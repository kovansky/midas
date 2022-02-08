package midas

type Site struct {
	SiteName string `json:"siteName"`
	RootDir  string `json:"rootDir"`
	Service  string `json:"service"`
	Registry struct {
		Type     string `json:"type"`
		Location string `json:"location"`
	} `json:"registry"`
	CollectionTypes map[string]Model `json:"collectionTypes"`
	SingleTypes     map[string]Model `json:"singleTypes"`
}

type Model struct {
	ArchetypePath string `json:"archetypePath,omitempty"`
	OutputDir     string `json:"outputDir,omitempty"`
}

type SiteService interface {
	GetRegistryService() (RegistryService, error)
	BuildSite(useCache bool) error
	CreateEntry(payload Payload) (string, error)
	UpdateEntry(payload Payload) (string, error)
	DeleteEntry(payload Payload) (string, error)
}
