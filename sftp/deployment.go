/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package sftp

import (
	"fmt"
	"github.com/kovansky/midas"
	"github.com/kovansky/midas/walk"
	"os"
	"path/filepath"
)

var _ midas.Deployment = (*Deployment)(nil)

type Deployment struct {
	site               midas.Site
	deploymentSettings midas.DeploymentSettings
	publicPath         string

	sftpClient Client
}

func New(site midas.Site, deploymentSettings midas.DeploymentSettings) (midas.Deployment, error) {
	// Get build destination directory
	var publicPath string
	if site.OutputSettings.Build != "" {
		if filepath.IsAbs(site.OutputSettings.Build) {
			publicPath = site.OutputSettings.Build
		} else {
			publicPath = filepath.Join(site.RootDir, site.OutputSettings.Build)
		}
	} else {
		publicPath = filepath.Join(site.RootDir, "public")
	}

	sftpClient := *NewClient(deploymentSettings.SFTP)

	return &Deployment{
		site:               site,
		deploymentSettings: deploymentSettings,
		publicPath:         publicPath,

		sftpClient: sftpClient,
	}, nil
}

// Deploy uploads the built files to the remote SFTP server.
func (d *Deployment) Deploy() error {
	// Retrieve local files.
	walker, err := d.retrieveFiles()
	if err != nil {
		return err
	}

	// And get local files as file map
	fileMap, err := d.getFileMap(walker)
	if err != nil {
		return err
	}

	// Get remote files.
	remoteFiles, err := d.remoteFiles()

	// Generate diffs
	uploads, removals := fileMap.Diff(remoteFiles)

	return nil
}

// remoteFiles returns a map of remote files indexed by their relative path.
func (d *Deployment) remoteFiles() (walk.FileMap, error) {
	err := d.sftpClient.Connect()
	if err != nil {
		return nil, err
	}
	defer func(sftpClient *Client) {
		_ = sftpClient.Close()
	}(&d.sftpClient)

	files, errors := d.sftpClient.RemoteFiles()
	if errors != nil {
		var errorsString string
		for _, err := range errors {
			errorsString += err.Error() + "\n"
		}

		return nil, fmt.Errorf("errors getting remote files:%s\n", errorsString)
	}

	return files, nil
}

// retrieveFiles walks the public directory and returns a channel of files to be uploaded.
func (d *Deployment) retrieveFiles() (walk.FileWalk, error) {
	walker := make(walk.FileWalk)

	// Gather the files to upload by walking the path recursively.
	go func() {
		defer close(walker)
		if err := filepath.Walk(d.publicPath, walker.Walk); err != nil {
			panic(err)
		}
	}()

	return walker, nil
}

// getFileMap returns locally retreived files in form of a fileMap indexed by their relative path.
func (d *Deployment) getFileMap(fileWalk walk.FileWalk) (walk.FileMap, error) {
	fileMap := make(walk.FileMap)

	// Gather the files to upload by walking the path recursively.
	for file := range fileWalk {
		fileInfo, err := os.Stat(file)
		if err != nil {
			return nil, err
		}

		relPath, _ := filepath.Rel(d.publicPath, file)

		fileMap[relPath] = fileInfo
	}

	return fileMap, nil
}
