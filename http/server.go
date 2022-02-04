package http

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
	"github.com/kovansky/midas"
	"golang.org/x/crypto/acme/autocert"
	"net"
	"net/http"
	"strings"
	"time"
)

const shutdownTimeout = 1 * time.Second

type Server struct {
	listener net.Listener
	server   *http.Server
	router   *chi.Mux

	Config midas.Config

	SiteServices map[string]func(site midas.Site) midas.SiteService
}

func NewServer() *Server {
	s := &Server{
		server: &http.Server{},
		router: chi.NewRouter(),
	}

	logger := httplog.NewLogger("midas", httplog.Options{Concise: true})

	s.router.Use(httplog.RequestLogger(logger))
	s.router.Use(middleware.Heartbeat("/ping"))

	s.server.Handler = http.HandlerFunc(s.router.ServeHTTP)

	s.router.Route("/", func(router chi.Router) {
		router.Use(s.authenticate)

		// Register specific routes
		s.registerStrapiToHugoRoutes(router)
	})

	return s
}

// UseTLS returns true if the certificate and key file are specified.
func (s *Server) UseTLS() bool {
	return s.Config.Domain != ""
}

// Scheme returns the URL scheme for the server.
func (s *Server) Scheme() string {
	if s.UseTLS() {
		return "https"
	}

	return "http"
}

// Port returns the TCP port for the running server.
func (s *Server) Port() int {
	if s.listener == nil {
		return 0
	}

	return s.listener.Addr().(*net.TCPAddr).Port
}

func (s *Server) URL() string {
	scheme, port := s.Scheme(), s.Port()

	// Use localhost unless a domain is specified
	domain := "localhost"
	if s.Config.Domain != "" {
		domain = s.Config.Domain
	}

	// Return without port if standard ports are used.
	if (scheme == "http" && port == 80) || (scheme == "https" && port == 443) {
		return fmt.Sprintf("%s://%s", scheme, domain)
	}
	return fmt.Sprintf("%s://%s:%d", scheme, domain, port)
}

// ListenAndServeTLSRedirect runs an HTTP server on port 80 to redirect users
// to the TLS-enabled port 443 server.
func ListenAndServeTLSRedirect(domain string) error {
	return http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+domain, http.StatusFound)
	}))
}

// Open validates server options and begins listening on the bind address.
func (s *Server) Open() error {
	var err error

	// Open a listener on address
	if s.Config.Domain != "" {
		s.listener = autocert.NewListener(s.Config.Domain)
	} else {
		if s.listener, err = net.Listen("tcp", s.Config.Addr); err != nil {
			return err
		}
	}

	go s.server.Serve(s.listener)

	return nil
}

// Close gracefully shut downs the server
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	return s.server.Shutdown(ctx)
}

func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Login via API key
		if k := r.Header.Get("Authorization"); strings.HasPrefix(k, "Bearer ") {
			apiKey := strings.TrimPrefix(k, "Bearer ")

			// Lookup user by API key.
			cfg, ok := s.Config.Sites[apiKey]

			if !ok {
				Error(w, r, midas.Errorf(midas.ErrUnauthorized, "Invalid API key."))
				return
			}

			r = r.WithContext(midas.NewContextWithSiteConfig(r.Context(), cfg))

			next.ServeHTTP(w, r)
			return
		}

		Error(w, r, midas.Errorf(midas.ErrUnauthorized, "No API key."))
	})
}
