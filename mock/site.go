package mock

import "github.com/kovansky/midas"

type SiteService struct {
	GetRegistryFn    func() (string, error)
	CreateRegistryFn func() (string, error)
	BuildSiteFn      func(useCache bool) error
	CreateEntryFn    func(payload midas.Payload) (string, error)
	UpdateEntryFn    func(payload midas.Payload) (string, error)
	RemoveEntryFn    func(payload midas.Payload) (string, error)
}

func (s *SiteService) GetRegistry() (string, error) {
	return s.GetRegistryFn()
}

func (s *SiteService) CreateRegistry() (string, error) {
	return s.CreateRegistryFn()
}

func (s *SiteService) BuildSite(useCache bool) error {
	return s.BuildSiteFn(useCache)
}

func (s *SiteService) CreateEntry(payload midas.Payload) (string, error) {
	return s.CreateEntryFn(payload)
}

func (s *SiteService) UpdateEntry(payload midas.Payload) (string, error) {
	return s.UpdateEntryFn(payload)
}

func (s *SiteService) RemoveEntry(payload midas.Payload) (string, error) {
	return s.RemoveEntryFn(payload)
}
