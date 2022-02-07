package hugo

import (
	"errors"
	"fmt"
	"github.com/kovansky/midas"
	"html/template"
	"os"
	"os/exec"
	"path"
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
	var arg string

	if !useCache {
		arg = "--ignoreCache"
	} else {
		arg = ""
	}

	cmd := exec.Command("hugo", arg)
	cmd.Dir = s.Site.RootDir

	out, err := cmd.Output()
	if err != nil {
		return midas.Errorf(midas.ErrInternal, "hugo build errored: %s\ncommand output: %s", err, out)
	}

	return nil
}

func (s SiteService) CreateEntry(payload midas.Payload) (string, error) {
	modelName := payload.Metadata()["model"].(string)
	model, _ := s.getModel(modelName)
	archetypePath := model.ArchetypePath
	if !path.IsAbs(archetypePath) {
		archetypePath = path.Join(s.Site.RootDir, archetypePath)
	}
	outputDir := model.OutputDir
	if !path.IsAbs(outputDir) {
		outputDir = path.Join(s.Site.RootDir, outputDir)
	}

	if !fileExists(archetypePath) {
		return "", midas.Errorf(midas.ErrSiteConfig, "archetype for model %s does not exist", modelName)
	}
	if !fileExists(outputDir) {
		err := os.Mkdir(outputDir, 0775)
		if err != nil {
			return "", err
		}
	}

	title := fmt.Sprintf("%v", payload.Entry()["Title"])
	slug := midas.CreateSlug(title)
	outputPath := path.Join(outputDir, slug+".html")

	if fileExists(outputPath) {
		return "", midas.Errorf(midas.ErrInvalid, "output file %s already exists", path.Base(outputPath))
	}

	tmpl, err := template.ParseFiles(archetypePath)

	if err != nil {
		return "", err
	}

	output, err := os.Create(outputPath)
	defer func(output *os.File) {
		_ = output.Close()
	}(output)

	if err != nil {
		return "", err
	}

	err = tmpl.Execute(output, struct {
		Entry map[string]interface{}
	}{payload.Entry()})
	if err != nil {
		return "", err
	}

	return outputPath, nil
}

func (SiteService) UpdateEntry(payload midas.Payload) (string, error) {
	// TODO implement me
	panic("implement me")
}

func (SiteService) RemoveEntry(payload midas.Payload) (string, error) {
	// TODO implement me
	panic("implement me")
}

// getModel returns a model from any type (collection or single), and true if model is single or false otherwise.
func (s SiteService) getModel(model string) (*midas.Model, bool) {
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
