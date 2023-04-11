/*
 * Copyright (c) 2023.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package astro

import (
	"github.com/kovansky/midas"
	"os/exec"
)

var _ midas.SiteService = (*SiteService)(nil)

type SiteService struct {
	Site midas.Site

	registry midas.RegistryService
}

func NewSiteService(config midas.Site) (midas.SiteService, error) {
	if _, ok := midas.RegistryServices[config.Registry.Type]; !ok {
		return nil, midas.Errorf(midas.ErrSiteConfig, "requested registry type %s does not exit", config.Registry.Type)
	}

	siteService := SiteService{
		Site:     config,
		registry: midas.RegistryServices[config.Registry.Type](config),
	}

	err := siteService.registry.OpenStorage()
	if err != nil {
		err = siteService.registry.CreateStorage()
		if err != nil {
			return nil, err
		}
	}

	return siteService, nil
}

func (s SiteService) GetRegistryService() (midas.RegistryService, error) {
	return s.registry, nil
}

func (s SiteService) BuildSite(_ bool) error {
	cmd := exec.Command("astro", "build")
	cmd.Dir = s.Site.RootDir

	out, err := cmd.CombinedOutput()

	if err != nil {
		return midas.Errorf(midas.ErrInternal, "astro build errored: %s\ncommand output: %s", err, out)
	}

	return nil
}

// ToDo: implement.
// As for now only build process is needed, the project it is used for gets the data directly from API.

func (s SiteService) CreateEntry(_ midas.Payload) (string, error) {
	return "", nil
}

func (s SiteService) UpdateEntry(_ midas.Payload) (string, error) {
	return "", nil
}

func (s SiteService) DeleteEntry(_ midas.Payload) (string, error) {
	return "", nil
}

func (s SiteService) UpdateSingle(_ midas.Payload) (string, error) {
	return "", nil
}
