package mock

import (
	"github.com/kovansky/midas"
)

type RegistryService struct {
	OpenStorageFn   func() error
	CloseStorageFn  func()
	CreateStorageFn func() error
	RemoveStorageFn func() error
	FlushFn         func() error
	CreateEntryFn   func(id, filename string) error
	ReadEntryFn     func(id string) (string, error)
	UpdateEntryFn   func(id, newFilename string) error
	DeleteEntryFn   func(id string) error

	Site midas.Site
}

func NewRegistryService(_ midas.Site) *RegistryService {
	return &RegistryService{}
}

func (r *RegistryService) OpenStorage() error {
	return r.OpenStorageFn()
}

func (r *RegistryService) CloseStorage() {
	r.CloseStorageFn()
}

func (r *RegistryService) CreateStorage() error {
	return r.CreateStorageFn()
}

func (r *RegistryService) RemoveStorage() error {
	return r.RemoveStorageFn()
}

func (r *RegistryService) Flush() error {
	return r.FlushFn()
}

func (r *RegistryService) CreateEntry(id, filename string) error {
	return r.CreateEntryFn(id, filename)
}

func (r *RegistryService) ReadEntry(id string) (string, error) {
	return r.ReadEntryFn(id)
}

func (r *RegistryService) UpdateEntry(id, newFilename string) error {
	return r.UpdateEntryFn(id, newFilename)
}

func (r *RegistryService) DeleteEntry(id string) error {
	return r.DeleteEntryFn(id)
}
