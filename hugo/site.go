package hugo

import "github.com/kovansky/strapi2hugo"

var _ strapi2hugo.SiteService = (*SiteService)(nil)

type SiteService struct {
	Site strapi2hugo.Site
}

func NewSiteService(config interface{}) strapi2hugo.SiteService {
	//TODO implement me
	panic("implement me")
}

func (SiteService) GetRegistry() (string, error) {
	//TODO implement me
	panic("implement me")
}

func (SiteService) CreateRegistry() (string, error) {
	//TODO implement me
	panic("implement me")
}

func (SiteService) BuildSite(useCache bool) error {
	//TODO implement me
	panic("implement me")
}

func (SiteService) CreateEntry(payload interface{}) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (SiteService) UpdateEntry(payload interface{}) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (SiteService) RemoveEntry(payload interface{}) (string, error) {
	//TODO implement me
	panic("implement me")
}
