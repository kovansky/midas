package hugo

import (
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
}

func NewSiteService(config midas.Site) midas.SiteService {
	return SiteService{Site: config}
}

func (SiteService) GetRegistry() (string, error) {
	// TODO implement me
	panic("implement me")
}

func (SiteService) CreateRegistry() (string, error) {
	// TODO implement me
	panic("implement me")
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
	archetypesDir := path.Join(s.Site.RootDir, "archetypes")
	defaultArchetype := path.Join(archetypesDir, "default.md")
	outputDir := path.Join(s.Site.RootDir, "content", payload.Metadata()["model"].(string)+"s")
	title := fmt.Sprintf("%v", payload.Entry()["Title"])
	slug := midas.CreateSlug(title)
	outputPath := path.Join(outputDir, slug+".html")

	tmpl, err := template.ParseFiles(defaultArchetype)

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
