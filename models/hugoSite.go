package models

import (
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path"
)

type HugoSite struct {
	SiteName string `json:"siteName"`
	RootDir  string `json:"rootDir"`
}

func (hugo HugoSite) CreateEntry(payload WebhookPayload) bool {
	archetypesDir := path.Join(hugo.RootDir, "archetypes")
	defaultArchetype := path.Join(archetypesDir, "default.md")
	outputDir := path.Join(hugo.RootDir, "content", payload.Model+"s")
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

func (hugo HugoSite) RebuildSite() bool {
	cmd := exec.Command("hugo")
	cmd.Dir = hugo.RootDir

	out, err := cmd.Output()
	if err != nil {
		fmt.Println(string(out))
		return false
	}

	fmt.Println(string(out))
	return true
}
