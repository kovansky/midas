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
	BuildSiteCounter   = 0
	CreateEntryCounter = 0
)

func resetCounters() {
	BuildSiteCounter = 0
	CreateEntryCounter = 0
}

// Server represents a test wrapper for midashttp.Server.
// It attaches mocks to the server & initializes on random port.
type Server struct {
	*midashttp.Server
}

// MustOpenServer is a test helper function for starting a new test HTTP server.
// Fail on error.
func MustOpenServer(tb testing.TB, siteServices map[string]func(site midas.Site) midas.SiteService, config midas.Config) *Server {
	tb.Helper()

	midas.Commit = "testing"
	midas.Version = "testing"

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
	s := MustOpenServer(t, map[string]func(site midas.Site) midas.SiteService{
		"hugo": func(_ midas.Site) midas.SiteService {
			siteService := mock.NewSiteService()

			siteService.BuildSiteFn = func(useCache bool) error {
				BuildSiteCounter++

				return nil
			}

			siteService.CreateEntryFn = func(_ midas.Payload) (string, error) {
				CreateEntryCounter++

				return "", nil
			}

			return siteService
		},
	}, midas.Config{
		Sites: map[string]midas.Site{
			"test": {
				Service:         "hugo",
				CollectionTypes: []string{"post"},
				SingleTypes:     []string{"homepage"},
			},
			"otherService": {
				Service: "other",
			},
		},
	})

	return s
}
