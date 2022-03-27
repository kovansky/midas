/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package http_test

import (
	"context"
	"github.com/kovansky/midas"
	midashttp "github.com/kovansky/midas/http"
	"github.com/kovansky/midas/mock"
	"io"
	"net/http"
	"testing"
)

var (
	MockSiteCounters = map[string]int{
		"GetRegistryService": 0,
		"BuildSite":          0,
		"CreateEntry":        0,
		"UpdateEntry":        0,
		"DeleteEntry":        0,
		"UpdateSingle":       0,
	}
	MockRegistryCounters = map[string]int{
		"OpenStorage":   0,
		"CloseStorage":  0,
		"CreateStorage": 0,
		"RemoveStorage": 0,
		"Flush":         0,
		"CreateEntry":   0,
		"ReadEntry":     0,
		"UpdateEntry":   0,
		"DeleteEntry":   0,
	}
)

func resetCounters() {
	for key := range MockSiteCounters {
		MockSiteCounters[key] = 0
	}

	for key := range MockRegistryCounters {
		MockRegistryCounters[key] = 0
	}
}

// Server represents a test wrapper for midashttp.Server.
// It attaches mocks to the server & initializes on random port.
type Server struct {
	*midashttp.Server
}

// MustOpenServer is a test helper function for starting a new test HTTP server.
// Fail on error.
func MustOpenServer(tb testing.TB, siteServices map[string]func(site midas.Site) (midas.SiteService, error), config midas.Config) *Server {
	tb.Helper()

	midas.Commit = "testing"
	midas.Version = "testing"

	midas.RegistryServices = map[string]func(site midas.Site) midas.RegistryService{
		"mock": prepareMockRegistryService,
	}

	// Init wrapper and set test config settings.
	s := &Server{Server: midashttp.NewServer(true)}

	s.SiteServices = siteServices
	s.Config = config

	if err := s.Open(); err != nil {
		tb.Fatal(err)
	}

	return s
}

// MustCloseServer is a test helper function for shutting down the server.
// Fail on error.
func MustCloseServer(tb testing.TB, s *Server) {
	tb.Helper()

	if err := s.Close(); err != nil {
		tb.Fatal(err)
	}
}

// MustNewRequest creates a new HTTP request using server's base URL.
func (s *Server) MustNewRequest(tb testing.TB, _ context.Context, apiKey, method, url string, body io.Reader) *http.Request {
	tb.Helper()

	r, err := http.NewRequest(method, s.URL()+url, body)
	if err != nil {
		tb.Fatal(err)
	}

	if apiKey != "" {
		r.Header.Set("Authorization", "Bearer "+apiKey)
	}

	return r
}

func SetUp(t *testing.T) *Server {
	s := MustOpenServer(t, map[string]func(site midas.Site) (midas.SiteService, error){
		"hugo": func(site midas.Site) (midas.SiteService, error) {
			siteService := mock.NewSiteService()

			siteService.BuildSiteFn = func(useCache bool) error {
				MockSiteCounters["BuildSite"]++

				return nil
			}
			siteService.CreateEntryFn = func(_ midas.Payload) (string, error) {
				MockSiteCounters["CreateEntry"]++

				return "", nil
			}
			siteService.UpdateEntryFn = func(_ midas.Payload) (string, error) {
				MockSiteCounters["UpdateEntry"]++

				return "", nil
			}
			siteService.DeleteEntryFn = func(_ midas.Payload) (string, error) {
				MockSiteCounters["DeleteEntry"]++

				return "", nil
			}
			siteService.GetRegistryServiceFn = func() (midas.RegistryService, error) {
				MockSiteCounters["GetRegistryService"]++

				return prepareMockRegistryService(site), nil
			}
			siteService.UpdateSingleFn = func(_ midas.Payload) (string, error) {
				MockSiteCounters["UpdateSingle"]++

				return "", nil
			}

			return siteService, nil
		},
	}, midas.Config{
		Sites: map[string]midas.Site{
			"test": {
				Service: "hugo",
				Registry: midas.RegistrySettings{
					Type: "mock",
				},
				CollectionTypes: map[string]midas.ModelSettings{
					"post": {"./archetypes/archetype.md", "./out"},
				},
				SingleTypes: map[string]midas.ModelSettings{
					"homepage": {},
				},
			},
			"otherService": {
				Service: "other",
			},
		},
	})

	return s
}

func prepareMockRegistryService(site midas.Site) midas.RegistryService {
	registryService := mock.NewRegistryService(site)

	registryService.OpenStorageFn = func() error {
		MockRegistryCounters["OpenStorage"]++

		return nil
	}
	registryService.CloseStorageFn = func() {
		MockRegistryCounters["CloseStorage"]++
	}
	registryService.CreateStorageFn = func() error {
		MockRegistryCounters["CreateStorage"]++

		return nil
	}
	registryService.FlushFn = func() error {
		MockRegistryCounters["Flush"]++

		return nil
	}
	registryService.CreateEntryFn = func(id, _ string) error {
		MockRegistryCounters["CreateEntry"]++

		if id == "error" {
			return midas.Errorf(midas.ErrRegistry, "entry already exists")
		}

		return nil
	}
	registryService.ReadEntryFn = func(id string) (string, error) {
		MockRegistryCounters["ReadEntry"]++

		if id == "error" {
			return "", midas.Errorf(midas.ErrRegistry, "entry doesn't exist")
		}

		return id + ".html", nil
	}
	registryService.UpdateEntryFn = func(id, _ string) error {
		MockRegistryCounters["UpdateEntry"]++

		if id == "error" {
			return midas.Errorf(midas.ErrRegistry, "entry doesn't exist")
		}

		return nil
	}
	registryService.DeleteEntryFn = func(id string) error {
		MockRegistryCounters["DeleteEntry"]++

		if id == "error" {
			return midas.Errorf(midas.ErrRegistry, "entry doesn't exist")
		}

		return nil
	}

	return registryService
}
