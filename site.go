package strapi2hugo

type Site struct {
	SiteName        string   `json:"siteName"`
	RootDir         string   `json:"rootDir"`
	CollectionTypes []string `json:"collectionTypes"`
	SingleTypes     []string `json:"singleTypes"`
}

type SiteService interface {
	GetRegistry() (string, error)
	CreateRegistry() (string, error)
	BuildSite(useCache bool) error
	CreateEntry(payload Payload) (string, error)
	UpdateEntry(payload Payload) (string, error)
	RemoveEntry(payload Payload) (string, error)
}
