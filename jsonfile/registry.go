/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package jsonfile

import (
	"encoding/json"
	"github.com/kovansky/midas"
	"io"
	"os"
	"path/filepath"
)

type RegistryService struct {
	path     string
	file     *os.File
	registry midas.Registry

	Site midas.Site
}

func NewRegistryService(site midas.Site) midas.RegistryService {
	filePath := filepath.Clean(site.Registry.Location)
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(site.RootDir, site.Registry.Location)
	}

	return &RegistryService{
		path: filePath,
		Site: site,
	}
}

// OpenStorage opens the registry file (and creates it if it doesn't exist) and then
// unmarshals the file content into the registry.
func (r *RegistryService) OpenStorage() error {
	file, err := os.OpenFile(r.path, os.O_RDWR|os.O_CREATE, 0775)
	if err != nil {
		return err
	}

	r.file = file

	if err = r.readStorage(); err != nil {
		return err
	}

	return nil
}

// readStorage reads the file content and unmarshals it into the registry.
func (r *RegistryService) readStorage() error {
	// Move cursor to the beginning of file
	if _, err := r.file.Seek(0, 0); err != nil {
		return err
	}

	data, err := io.ReadAll(r.file)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		r.registry = make(map[string]string)

		return nil
	}

	err = json.Unmarshal(data, &r.registry)
	if err != nil {
		return err
	}

	return nil
}

// CloseStorage closes the file handler.
func (r *RegistryService) CloseStorage() {
	_ = r.file.Close()
}

// CreateStorage creates the registry file. Equivalent of calling OpenStorage in this case.
func (r *RegistryService) CreateStorage() error {
	return r.OpenStorage()
}

// RemoveStorage closes file handle and removes the registry file.
func (r *RegistryService) RemoveStorage() error {
	r.CloseStorage()
	err := os.Remove(r.path)
	if err != nil {
		return err
	}

	return nil
}

// Flush writes the working changes on registry to the file.
func (r *RegistryService) Flush() error {
	// Marshal the Registry into JSON
	content, err := json.MarshalIndent(r.registry, "", "\t")
	if err != nil {
		return err
	}

	// Remove file contents
	if err = r.file.Truncate(0); err != nil {
		return err
	}
	// Move cursor to the beginning of file
	if _, err = r.file.Seek(0, 0); err != nil {
		return err
	}

	// Write JSON to file
	if _, err = r.file.Write(content); err != nil {
		return err
	}

	return nil
}

// CreateEntry appends a new id to filename mapping to the registry.
func (r *RegistryService) CreateEntry(id, filename string) error {
	if _, err := r.ReadEntry(id); err == nil {
		return midas.Errorf(midas.ErrRegistry, "entry %s already exists", id)
	}

	r.registry[id] = filename
	return nil
}

// ReadEntry returns filename attached to given id from the registry.
func (r *RegistryService) ReadEntry(id string) (string, error) {
	if _, ok := r.registry[id]; !ok {
		return "", midas.Errorf(midas.ErrRegistry, "entry %s doesn't exist", id)
	}

	return r.registry[id], nil
}

// UpdateEntry sets a new filename for the id in the registry.
func (r *RegistryService) UpdateEntry(id, newFilename string) error {
	if _, err := r.ReadEntry(id); err != nil {
		return midas.Errorf(midas.ErrRegistry, "entry %s doesn't exist", id)
	}

	r.registry[id] = newFilename
	return nil
}

// DeleteEntry removes entry with given id from the registry.
func (r *RegistryService) DeleteEntry(id string) error {
	if _, err := r.ReadEntry(id); err != nil {
		return midas.Errorf(midas.ErrRegistry, "entry %s doesn't exist", id)
	}

	delete(r.registry, id)
	return nil
}
