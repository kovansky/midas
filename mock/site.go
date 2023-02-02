/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package mock

import (
	"github.com/kovansky/midas"
	"github.com/rs/zerolog"
)

var _ midas.SiteService = (*SiteService)(nil)

type SiteService struct {
	GetRegistryServiceFn func() (midas.RegistryService, error)
	CreateRegistryFn     func() (string, error)
	BuildSiteFn          func(useCache bool, log zerolog.Logger) error
	CreateEntryFn        func(payload midas.Payload) (string, error)
	UpdateEntryFn        func(payload midas.Payload) (string, error)
	DeleteEntryFn        func(payload midas.Payload) (string, error)
	UpdateSingleFn       func(payload midas.Payload) (string, error)
}

func NewSiteService() *SiteService {
	return &SiteService{}
}

func (s *SiteService) GetRegistryService() (midas.RegistryService, error) {
	return s.GetRegistryServiceFn()
}

func (s *SiteService) BuildSite(useCache bool, log zerolog.Logger) error {
	return s.BuildSiteFn(useCache, log)
}

func (s *SiteService) CreateEntry(payload midas.Payload) (string, error) {
	return s.CreateEntryFn(payload)
}

func (s *SiteService) UpdateEntry(payload midas.Payload) (string, error) {
	return s.UpdateEntryFn(payload)
}

func (s *SiteService) DeleteEntry(payload midas.Payload) (string, error) {
	return s.DeleteEntryFn(payload)
}

func (s *SiteService) UpdateSingle(payload midas.Payload) (string, error) {
	return s.UpdateSingleFn(payload)
}
