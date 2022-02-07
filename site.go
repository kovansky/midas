package midas

type Site struct {
	SiteName string `json:"siteName"`
	RootDir  string `json:"rootDir"`
	Service  string `json:"service"`
	Registry struct {
		Type     string `json:"type"`
		Location string `json:"location"`
	} `json:"registry"`
	CollectionTypes []string `json:"collectionTypes"`
	SingleTypes     []string `json:"singleTypes"`
}

type SiteService interface {
	GetRegistryService() (RegistryService, error)
	BuildSite(useCache bool) error
	CreateEntry(payload Payload) (string, error)
	UpdateEntry(payload Payload) (string, error)
	RemoveEntry(payload Payload) (string, error)
}
