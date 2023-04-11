/*
 * Copyright (c) 2023.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package none

import "github.com/kovansky/midas"

// RegistryService in none package is a "dummy" registry service when no actual read/write features are needed,
// i.e. in case we do not create or manage any files locally.
type RegistryService struct {
}

func NewRegistryService(_ midas.Site) midas.RegistryService {
	return &RegistryService{}
}

func (r RegistryService) OpenStorage() error {
	return nil
}

func (r RegistryService) CloseStorage() {
	return
}

func (r RegistryService) CreateStorage() error {
	return nil
}

func (r RegistryService) RemoveStorage() error {
	return nil
}

func (r RegistryService) Flush() error {
	return nil
}

func (r RegistryService) CreateEntry(_, _ string) error {
	return nil
}

func (r RegistryService) ReadEntry(_ string) (string, error) {
	return "", nil
}

func (r RegistryService) UpdateEntry(_, _ string) error {
	return nil
}

func (r RegistryService) DeleteEntry(_ string) error {
	return nil
}
