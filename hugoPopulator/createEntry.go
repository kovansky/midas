package hugoPopulator

import (
	"fmt"
	"github.com/kovansky/strapi2hugo/models"
	"html/template"
	"os"
	"path"
)

func (hugo HugoSite) CreateEntry(payload models.WebhookPayload) bool {
	archetypesDir := path.Join(hugo.rootDir, "archetypes")
	defaultArchetype := path.Join(archetypesDir, "default.md")
	outputDir := path.Join(hugo.rootDir, "content", payload.Model+"s")
	title := fmt.Sprintf("%v", payload.Entry["Title"])
	slug := createSlug(title)
	outputPath := path.Join(outputDir, slug+".html")

	tmpl, err := template.ParseFiles(defaultArchetype)

	if err != nil {
		fmt.Println(err)
		return false
	}

	output, err := os.Create(outputPath)
	defer func(output *os.File) {
		_ = output.Close()
	}(output)

	if err != nil {
		fmt.Println(err)
		return false
	}

	err = tmpl.Execute(output, struct {
		Entry map[string]interface{}
	}{payload.Entry})
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}
