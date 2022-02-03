package http_test

import (
	"context"
	"github.com/kovansky/midas"
	midashttp "github.com/kovansky/midas/http"
	"io"
	"net/http"
	"testing"
)

// Server represents a test wrapper for midashttp.Server.
// It attaches mocks to the server & initializes on random port.
type Server struct {
	*midashttp.Server
}

// MustOpenServer is a test helper function for starting a new test HTTP server.
// Fail on error.
func MustOpenServer(tb testing.TB, siteServices map[string]func(site midas.Site) midas.SiteService, config midas.Config) *Server {
	tb.Helper()

	// Init wrapper and set test config settings.
	s := &Server{Server: midashttp.NewServer()}

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
