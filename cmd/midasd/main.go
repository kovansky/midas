/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kovansky/midas"
	"github.com/kovansky/midas/astro"
	"github.com/kovansky/midas/aws"
	"github.com/kovansky/midas/bluemonday"
	"github.com/kovansky/midas/concurrent"
	"github.com/kovansky/midas/http"
	"github.com/kovansky/midas/hugo"
	"github.com/kovansky/midas/jsonfile"
	"github.com/kovansky/midas/none"
	"github.com/kovansky/midas/sftp"
	"github.com/rollbar/rollbar-go"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
)

var (
	commit, version, date, environment, logLevel string
	logLevelAcceptedValues                       = map[string]struct{}{
		"trace":    {},
		"debug":    {},
		"info":     {},
		"warn":     {},
		"error":    {},
		"critical": {},
	}
)

// main is the entry point of the application binary.
func main() {
	// Propagate build info
	midas.Commit = commit
	midas.Version = version

	// Setup signal handlers
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; cancel() }()

	// Create a new type to represent the application.
	m := NewMain()

	// Parse command line flags and load the configuration.
	if err := m.ParseFlags(ctx, os.Args[1:]); err == flag.ErrHelp {
		os.Exit(1)
	} else if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Check if environment wasn't set in env variables (instead of flag)
	if os.Getenv("MIDAS_ENV") != "" {
		environment = os.Getenv("MIDAS_ENV")
	}

	// Execute program
	if err := m.Run(ctx); err != nil {
		_ = m.Close()
		_, _ = fmt.Fprintln(os.Stderr, err)
		midas.ReportError(ctx, err)
		os.Exit(1)
	}

	<-ctx.Done()

	if err := m.Close(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Main represents the program
type Main struct {
	// Configuration path and parsed config data
	Config     midas.Config
	ConfigPath string

	// HTTP server for handling HTTP communication
	HTTPServer *http.Server
}

func readConfig(filename string) (midas.Config, error) {
	config := defaultConfig()

	if buf, err := ioutil.ReadFile(filename); err != nil {
		return config, err
	} else if err := json.Unmarshal(buf, &config); err != nil {
		return config, err
	}

	return config, nil
}

func defaultConfig() midas.Config {
	return midas.Config{
		Addr: "127.0.0.1:8443",
	}
}

const (
	defaultConfigPath = "./config.json"
)

func NewMain() *Main {
	return &Main{
		Config:     defaultConfig(),
		ConfigPath: defaultConfigPath,

		HTTPServer: http.NewServer(logLevel, false),
	}
}

// Close gracefully stops the program
func (m *Main) Close() error {
	if m.HTTPServer != nil {
		return m.HTTPServer.Close()
	}

	return nil
}

// ParseFlags parses command line arguments and loads the config.
func (m *Main) ParseFlags(_ context.Context, args []string) error {
	// Only config path flag
	fs := flag.NewFlagSet("midasd", flag.ContinueOnError)
	fs.StringVar(&m.ConfigPath, "config", defaultConfigPath, "config path")
	fs.StringVar(&environment, "env", "production", "app environment (development, production)")
	fs.StringVar(&logLevel, "log", "info", "log level (trace, debug, info, warn, error, critical)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if _, ok := logLevelAcceptedValues[logLevel]; !ok {
		logLevel = "info"
	}

	configPath, err := expand(m.ConfigPath)
	if err != nil {
		return err
	}

	config, err := readConfig(configPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", m.ConfigPath)
	} else if err != nil {
		return err
	}

	m.Config = config

	return nil
}

// Run executes the program. The configuration should already be set up
// before calling this function.
func (m *Main) Run(_ context.Context) (err error) {
	log.Printf("Starting midas v%s (%s) built on %s\n", midas.Version, midas.Commit, date)
	log.Printf("Environment: %s\n", environment)

	if m.Config.RollbarToken != "" {
		rollbar.SetToken(m.Config.RollbarToken)
		rollbar.SetEnvironment(environment)
		rollbar.SetCodeVersion(version)
		rollbar.SetServerRoot("github.com/kovansky/midas")

		midas.ReportError = rollbarError

		log.Println("rollbar error tracking enabled")
	}

	m.HTTPServer.Config = m.Config

	m.HTTPServer.SiteServices = map[string]func(site midas.Site) (midas.SiteService, error){
		"hugo": func(site midas.Site) (midas.SiteService, error) {
			return hugo.NewSiteService(site)
		},
		"astro": func(site midas.Site) (midas.SiteService, error) {
			return astro.NewSiteService(site)
		},
	}

	midas.RegistryServices = map[string]func(site midas.Site) midas.RegistryService{
		"jsonfile": func(site midas.Site) midas.RegistryService {
			return jsonfile.NewRegistryService(site)
		},
		"none": func(site midas.Site) midas.RegistryService {
			return none.NewRegistryService(site)
		},
	}

	midas.DeploymentTargets = map[string]func(site midas.Site, settings midas.DeploymentSettings, isDraft bool) (midas.Deployment, error){
		"aws": func(site midas.Site, settings midas.DeploymentSettings, isDraft bool) (midas.Deployment, error) {
			return aws.New(site, settings, isDraft)
		},
		"sftp": func(site midas.Site, settings midas.DeploymentSettings, isDraft bool) (midas.Deployment, error) {
			return sftp.New(site, settings, isDraft)
		},
	}

	midas.Sanitizer = bluemonday.NewSanitizerService()

	midas.Concurrents = concurrent.NewList()

	if err := m.HTTPServer.Open(); err != nil {
		return err
	}

	if m.HTTPServer.UseTLS() {
		go func() {
			log.Fatal(http.ListenAndServeTLSRedirect(m.Config.Domain))
		}()
	}

	log.Printf("Running on %s", m.HTTPServer.URL())

	return nil
}

// expand changes tilde in path to user's home directory.
func expand(path string) (string, error) {
	// Ignore if path has no leading tilde
	if path != "~" && !strings.HasPrefix(path, "~"+string(os.PathSeparator)) {
		return path, nil
	}

	// Get current user to determine home path
	usr, err := user.Current()
	if err != nil {
		return path, err
	} else if usr.HomeDir == "" {
		return path, fmt.Errorf("home directory unset")
	}

	if path == "~" {
		return usr.HomeDir, nil
	}

	return filepath.Join(usr.HomeDir, strings.TrimPrefix(path, "~"+string(os.PathSeparator))), nil
}

// rollbarError reports internal errors to rollbar.
func rollbarError(ctx context.Context, err error, args ...interface{}) {
	if midas.ErrorCode(err) != midas.ErrInternal {
		return
	}

	if key := midas.ApiKeyFromContext(ctx); key != "" {
		rollbar.SetPerson(key, "", "")
	} else {
		rollbar.ClearPerson()
	}

	if len(args) > 0 {
		rollbar.Error(append([]interface{}{err}, args...)...)
	} else {
		rollbar.Error(err)
	}
}

// rollbarPanic reports panics to rollbar. Maybe will be used in future.
func _(err interface{}) {
	log.Printf("panic: %+v\n", err)
	rollbar.LogPanic(err, true)
}
