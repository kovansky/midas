package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kovansky/midas"
	"github.com/kovansky/midas/http"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
)

// main is the entry point of the application binary.
func main() {
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

	// Execute program
	if err := m.Run(ctx); err != nil {
		_ = m.Close()
		_, _ = fmt.Fprintln(os.Stderr, err)
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

		HTTPServer: http.NewServer(),
	}
}

// Close gracefully stops the program
func (m *Main) Close() error {
	if m.HTTPServer != nil {
		if err := m.HTTPServer.Close(); err != nil {
			return err
		}
	}

	return nil
}

// ParseFlags parses command line arguments and loads the config.
func (m *Main) ParseFlags(_ context.Context, args []string) error {
	// Only config path flag
	fs := flag.NewFlagSet("midasd", flag.ContinueOnError)
	fs.StringVar(&m.ConfigPath, "config", defaultConfigPath, "config path")

	if err := fs.Parse(args); err != nil {
		return err
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
	m.HTTPServer.Config = m.Config

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
