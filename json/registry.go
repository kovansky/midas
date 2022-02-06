package json

import (
	"encoding/json"
	"github.com/kovansky/midas"
	"io"
	"os"
	"path"
)

type RegistryService struct {
	filename string
	filePath string
	file     *os.File
	registry midas.Registry

	Site midas.Site
}

func NewRegistryService(site midas.Site) *RegistryService {
	filename := "midas-registry.json"

	return &RegistryService{
		filename: filename,
		filePath: path.Join(site.RootDir, filename),
		Site:     site,
	}
}

// OpenStorage opens the registry file (and creates it if doesn't exist) and then
// unmarshals the file content into the registry.
func (r *RegistryService) OpenStorage() error {
	file, err := os.OpenFile(r.filePath, os.O_RDWR|os.O_CREATE, 0775)
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
	err := os.Remove(r.filePath)
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
	r.registry[id] = filename
	return nil
}

// ReadEntry returns filename attached to given id from the registry.
func (r *RegistryService) ReadEntry(id string) (string, error) {
	return r.registry[id], nil
}

// UpdateEntry sets a new filename for the id in the registry.
// Equivalent of calling CreateEntry in this case.
func (r *RegistryService) UpdateEntry(id, newFilename string) error {
	return r.CreateEntry(id, newFilename)
}

// DeleteEntry removes entry with given id from the registry.
func (r *RegistryService) DeleteEntry(id string) error {
	delete(r.registry, r.registry[id])
	return nil
}
