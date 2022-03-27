/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package midas

type Site struct {
	SiteName        string                   `json:"siteName"`
	RootDir         string                   `json:"rootDir"`
	OutputSettings  OutputSettings           `json:"outputSettings"`
	BuildDrafts     bool                     `json:"buildDrafts"`
	DraftsUrl       string                   `json:"draftsUrl"`
	Service         string                   `json:"service"`
	Registry        RegistrySettings         `json:"registry"`
	CollectionTypes map[string]ModelSettings `json:"collectionTypes"`
	SingleTypes     map[string]ModelSettings `json:"singleTypes"`
}

type OutputSettings struct {
	Build string `json:"build,omitempty"`
	Draft string `json:"draft,omitempty"`
}

type ModelSettings struct {
	ArchetypePath string `json:"archetypePath,omitempty"`
	OutputDir     string `json:"outputDir,omitempty"`
}

type RegistrySettings struct {
	Type     string `json:"type"`
	Location string `json:"location"`
}

type SiteService interface {
	GetRegistryService() (RegistryService, error)
	BuildSite(useCache bool) error
	CreateEntry(payload Payload) (string, error)
	UpdateEntry(payload Payload) (string, error)
	DeleteEntry(payload Payload) (string, error)
	UpdateSingle(payload Payload) (string, error)
}
