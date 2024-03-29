/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package midas

import "github.com/rs/zerolog"

type Site struct {
	SiteName string `json:"siteName"`
	Service  string `json:"service"`

	RootDir        string         `json:"rootDir"`
	OutputSettings OutputSettings `json:"outputSettings"`

	BuildDrafts bool   `json:"buildDrafts,default=false"`
	DraftsUrl   string `json:"draftsUrl"`

	Registry        RegistrySettings         `json:"registry"`
	CollectionTypes map[string]ModelSettings `json:"collectionTypes"`
	SingleTypes     map[string]ModelSettings `json:"singleTypes"`

	Deployment       DeploymentSettings `json:"deployment"`
	DraftsDeployment DeploymentSettings `json:"draftsDeployment"`
}

type OutputSettings struct {
	Build            string `json:"build,omitempty"`
	Draft            string `json:"draft,omitempty"`
	DraftEnvironment string `json:"draftEnvironment,omitempty"`
}

type ModelSettings struct {
	ArchetypePath string `json:"archetypePath,omitempty"`
	OutputDir     string `json:"outputDir,omitempty"`
	Fields        struct {
		Title *string   `json:"title,omitempty"`
		HTML  *[]string `json:"html,omitempty"`
	} `json:"fields"`
}

type RegistrySettings struct {
	Type     string `json:"type"`
	Location string `json:"location"`
}

type SiteService interface {
	GetRegistryService() (RegistryService, error)
	BuildSite(useCache bool, log zerolog.Logger) error
	CreateEntry(payload Payload) (string, error)
	UpdateEntry(payload Payload) (string, error)
	DeleteEntry(payload Payload) (string, error)
	UpdateSingle(payload Payload) (string, error)
}
