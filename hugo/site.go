/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package hugo

import (
	"errors"
	"fmt"
	"github.com/kovansky/midas"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
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

func (s SiteService) BuildSite(useCache bool) error {
	var arg = s.constructBuildArgs(useCache, false)

	cmd := exec.Command("hugo", arg...)
	cmd.Dir = s.Site.RootDir

	out, err := cmd.Output()
	if err != nil {
		return midas.Errorf(midas.ErrInternal, "hugo build errored: %s\ncommand output: %s", err, out)
	}

	if s.Site.BuildDrafts {
		if err = s.BuildDrafts(); err != nil {
			return err
		}
	}

	return nil
}

// constructBuildArgs generates hugo build arguments. If `isDraft` is true, the destination is changed
// to draft destination and draft arguments are added.
func (s SiteService) constructBuildArgs(useCache, isDraft bool) (arg []string) {
	// In draft we never want to use cache to get the latest changes.
	if !useCache || isDraft {
		arg = append(arg, "--ignoreCache")
	}

	if !isDraft {
		if s.Site.OutputSettings.Build != "" {
			arg = append(arg, "-d", s.Site.OutputSettings.Build)
		}
	} else {
		arg = append(arg, "-d")

		if s.Site.OutputSettings.Draft != "" {
			arg = append(arg, s.Site.OutputSettings.Draft)
		} else {
			arg = append(arg, "publicDrafts")
		}

		// -D is for build drafts, -E for build expired, -F for build future
		arg = append(arg, "-D", "-E", "-F")

		// Add baseUrl, if specified
		if s.Site.DraftsUrl != "" {
			arg = append(arg, "-b", s.Site.DraftsUrl)
		}
	}

	return arg
}

func (s SiteService) BuildDrafts() error {
	var arg = s.constructBuildArgs(false, true)

	cmd := exec.Command("hugo", arg...)
	cmd.Dir = s.Site.RootDir

	out, err := cmd.Output()
	if err != nil {
		return midas.Errorf(midas.ErrInternal, "hugo draft build errored: %s\ncommand output: %s", err, out)
	}

	return nil
}

func (s SiteService) CreateEntry(payload midas.Payload) (string, error) {
	// Set archetype path and output directory
	modelName := payload.Metadata()["model"].(string)
	model, _ := s.getModel(modelName)
	archetypePath := model.ArchetypePath
	if !filepath.IsAbs(archetypePath) {
		archetypePath = filepath.Join(s.Site.RootDir, archetypePath)
	}
	outputDir := model.OutputDir
	if !filepath.IsAbs(outputDir) {
		outputDir = filepath.Join(s.Site.RootDir, outputDir)
	}

	// Check if archetype exists
	if !fileExists(archetypePath) {
		return "", midas.Errorf(midas.ErrSiteConfig, "archetype for model %s does not exist", modelName)
	}
	// Check if output dir exists, attempt to create it if it doesn't
	if !fileExists(outputDir) {
		err := os.Mkdir(outputDir, 0775)
		if err != nil {
			return "", err
		}
	}

	// Format output filename
	title := fmt.Sprintf("%v", payload.Entry()["Title"])
	slug := midas.CreateSlug(title)
	outputPath := filepath.Join(outputDir, slug+".html")

	// Check if output filename is free
	if fileExists(outputPath) {
		return "", midas.Errorf(midas.ErrInvalid, "output file %s already exists", filepath.Base(outputPath))
	}

	// Read archetype file
	tmpl, err := template.ParseFiles(archetypePath)
	if err != nil {
		return "", err
	}

	// Create output file
	output, err := os.Create(outputPath)
	defer func(output *os.File) {
		_ = output.Close()
	}(output)

	if err != nil {
		return "", err
	}

	// Parse archetype and write it to output
	err = executeTemplate(tmpl, output, payload)
	if err != nil {
		return "", err
	}

	// Add entry to registry
	entryId := s.EntryId(payload)

	if err = s.registry.CreateEntry(entryId, outputPath); err != nil {
		return outputPath, err
	}
	if err = s.registry.Flush(); err != nil {
		return outputPath, err
	}

	return outputPath, nil
}

func (s SiteService) UpdateEntry(payload midas.Payload) (string, error) {
	// Set archetype path
	modelName := payload.Metadata()["model"].(string)
	model, _ := s.getModel(modelName)
	archetypePath := model.ArchetypePath
	if !filepath.IsAbs(archetypePath) {
		archetypePath = filepath.Join(s.Site.RootDir, archetypePath)
	}

	// Check if archetype exists
	if !fileExists(archetypePath) {
		return "", midas.Errorf(midas.ErrSiteConfig, "archetype for model %s does not exist", modelName)
	}

	// Get old path
	entryId := s.EntryId(payload)
	oldPath, err := s.registry.ReadEntry(entryId)
	outputDir := filepath.Dir(oldPath)
	if err != nil {
		// If entry not in the registry, create empty one, otherwise UpdateEntry later will complain
		_ = s.registry.CreateEntry(entryId, "")
		// Read output dir in normal way
		outputDir = model.OutputDir
		if !filepath.IsAbs(outputDir) {
			outputDir = filepath.Join(s.Site.RootDir, outputDir)
		}
	}

	// Check if output dir exists, attempt to create it if it doesn't
	if !fileExists(outputDir) {
		err := os.Mkdir(outputDir, 0775)
		if err != nil {
			return "", err
		}
	}

	// Format new output filename
	title := fmt.Sprintf("%v", payload.Entry()["Title"])
	slug := midas.CreateSlug(title)
	outputPath := filepath.Join(outputDir, slug+".html")

	// Check if output filename is free (excluding situation where name doesn't changed)
	if fileExists(outputPath) && filepath.Base(outputPath) != filepath.Base(oldPath) {
		return "", midas.Errorf(midas.ErrInvalid, "output file %s already exists", filepath.Base(outputPath))
	}

	// Remove old entry if exists
	if oldPath != "" && fileExists(oldPath) {
		_ = os.Remove(oldPath)
	}

	// Read archetype file
	tmpl, err := template.ParseFiles(archetypePath)
	if err != nil {
		return "", err
	}

	// Create output file
	output, err := os.Create(outputPath)
	defer func(output *os.File) {
		_ = output.Close()
	}(output)

	if err != nil {
		return "", err
	}

	// Parse archetype and write it to output
	err = executeTemplate(tmpl, output, payload)
	if err != nil {
		return "", err
	}

	// Update entry in registry
	if err = s.registry.UpdateEntry(entryId, outputPath); err != nil {
		return outputPath, err
	}
	if err = s.registry.Flush(); err != nil {
		return outputPath, err
	}

	return outputPath, nil
}

func (s SiteService) DeleteEntry(payload midas.Payload) (string, error) {
	// Get entry path
	entryId := s.EntryId(payload)
	entryPath, err := s.registry.ReadEntry(entryId)
	if err != nil {
		return "", err
	}

	// Remove entry
	if err = os.Remove(entryPath); err != nil {
		return "", nil
	}

	// Remove entry from registry
	if err = s.registry.DeleteEntry(entryId); err != nil {
		return entryPath, err
	}
	if err = s.registry.Flush(); err != nil {
		return entryPath, err
	}

	return entryPath, nil
}

func (s SiteService) UpdateSingle(payload midas.Payload) (string, error) {
	// Set output directory
	modelName := payload.Metadata()["model"].(string)
	model, _ := s.getModel(modelName)
	outputDir := model.OutputDir
	if !filepath.IsAbs(outputDir) {
		outputDir = filepath.Join(s.Site.RootDir, outputDir)
	}

	// Check if output dir exists, attempt to create it if it doesn't
	if !fileExists(outputDir) {
		err := os.Mkdir(outputDir, 0775)
		if err != nil {
			return "", err
		}
	}

	// Format output filename
	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.json", modelName))

	// Sanitize the entry
	entry := payload.Entry()
	entry = sanitizeHtmlInMap(entry)

	payload.SetEntry(entry)

	asJson, err := payload.MarshalJSON()
	if err != nil {
		return "", err
	}

	// Open output file
	output, err := os.OpenFile(outputPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0775)
	defer func(outout *os.File) {
		_ = output.Close()
	}(output)

	if err != nil {
		return "", err
	}

	// Write the output
	_, err = output.Write(asJson)
	if err != nil {
		return "", err
	}

	return outputPath, nil
}

// EntryId generates the entry to be used in registry.
func (s SiteService) EntryId(payload midas.Payload) string {
	return fmt.Sprintf("%v-%v", payload.Metadata()["model"], payload.Entry()["id"])
}

// getModel returns a model from any type (collection or single), and true if model is single or false otherwise.
func (s SiteService) getModel(model string) (*midas.ModelSettings, bool) {
	if m, ok := s.Site.CollectionTypes[model]; ok {
		return &m, false
	} else if m, ok := s.Site.SingleTypes[model]; ok {
		return &m, true
	}

	return nil, true
}

// fileExists return true if path exists or false otherwise
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !errors.Is(err, os.ErrNotExist)
}

// executeTemplate sanitizes the HTML and executes the template to the output file
func executeTemplate(tmpl *template.Template, output *os.File, payload midas.Payload) (err error) {
	sanitized := payload.Entry()
	sanitized["Content"] = template.HTML(midas.Sanitizer.Sanitize(sanitized["Content"].(string)))

	// Parse archetype and write it to output
	err = tmpl.Execute(output, struct {
		Metadata map[string]interface{}
		Entry    map[string]interface{}
	}{payload.Metadata(), sanitized})
	if err != nil {
		return err
	}

	return nil
}

// sanitizeHtmlInMap iterates (recursively) through given map and passes each value through HTML sanitizer.
func sanitizeHtmlInMap(entry map[string]interface{}) map[string]interface{} {
	output := make(map[string]interface{})

	for key, value := range entry {
		// Check value type
		switch value.(type) {
		case map[string]interface{}:
			output[key] = sanitizeHtmlInMap(value.(map[string]interface{}))
			break
		case []interface{}:
			output[key] = sanitizeHtmlInSlice(value.([]interface{}))
			break
		default:
			if stringed, ok := value.(string); ok {
				output[key] = midas.Sanitizer.Sanitize(stringed)
			} else {
				output[key] = value
			}

		}
	}

	return output
}

// sanitizeHtmlInSlice iterates (recursively) through given slice and passes each value through HTML sanitizer.
func sanitizeHtmlInSlice(entry []interface{}) []interface{} {
	output := make([]interface{}, len(entry))

	for key, value := range entry {
		// Check value type
		switch value.(type) {
		case map[string]interface{}:
			output[key] = sanitizeHtmlInMap(value.(map[string]interface{}))
			break
		case []interface{}:
			output[key] = sanitizeHtmlInSlice(value.([]interface{}))
			break
		default:
			if stringed, ok := value.(string); ok {
				output[key] = midas.Sanitizer.Sanitize(stringed)
			} else {
				output[key] = value
			}
		}
	}

	return output
}
