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
	"github.com/kovansky/midas/aws"
	"github.com/kovansky/midas/bluemonday"
	"github.com/kovansky/midas/jsonfile"
	"github.com/kovansky/midas/sftp"
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
		midas.ReportError(ctx, err)
		os.Exit(1)
	}

	cancel()

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

	SFTPClient *sftp.Client
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
	}
}

// ParseFlags parses command line arguments and loads the config.
func (m *Main) ParseFlags(_ context.Context, args []string) error {
	// Only config path flag
	fs := flag.NewFlagSet("midasSshTest", flag.ContinueOnError)
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

// Close gracefully closes the program.
func (m *Main) Close() error {
	if m.SFTPClient != nil {
		return m.SFTPClient.Close()
	}

	return nil
}

// Run executes the program. The configuration should already be set up
// before calling this function.
func (m *Main) Run(_ context.Context) (err error) {
	log.Printf("Starting midas SSH Test\n")

	midas.RegistryServices = map[string]func(site midas.Site) midas.RegistryService{
		"jsonfile": func(site midas.Site) midas.RegistryService {
			return jsonfile.NewRegistryService(site)
		},
	}

	midas.DeploymentTargets = map[string]func(site midas.Site, settings midas.DeploymentSettings) (midas.Deployment, error){
		"aws": func(site midas.Site, settings midas.DeploymentSettings) (midas.Deployment, error) {
			return aws.New(site, settings)
		},
	}

	midas.Sanitizer = bluemonday.NewSanitizerService()

	var sftpClient *sftp.Client
	if site := mapFirstEntry(m.Config.Sites); site == nil {
		return fmt.Errorf("no sites configured")
	} else {
		sftpClient = sftp.NewClient(site.Deployment.SSH)
	}

	err = sftpClient.Connect()
	if err != nil {
		return err
	}

	m.SFTPClient = sftpClient

	files, errors := sftpClient.RemoteFiles()
	if errors != nil {
		var errorsString string
		for _, err := range errors {
			errorsString += err.Error() + "\n"
		}

		return fmt.Errorf("errors getting remote files:%s\n", errorsString)
	}

	for filePath, file := range files {
		log.Printf("%s: name - %s (is dir? %t), lastMod - %s", filePath, file.Name(), file.IsDir(), file.ModTime())
	}

	_ = sftpClient.Close()

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

// mapFirstEntry returns the first entry of the map or nil if the map is empty.
func mapFirstEntry[M ~map[K]V, K string, V any](m M) *V {
	for _, v := range m {
		return &v
	}
	return nil
}
