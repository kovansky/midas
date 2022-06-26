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
	"log"
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

func (d *Deployment) Deploy() error {
	// Retrieve files to upload.
	fileWalk, err := d.retrieveFiles()
	if err != nil {
		return err
	}

	// Get files as file map
	fileMap, err := d.getFileMap(fileWalk)
	if err != nil {
		return err
	}

	log.Println("Local")
	for filePath, file := range fileMap {
		log.Printf("%s: name - %s (is dir? %t), lastMod - %s", filePath, file.Name(), file.IsDir(), file.ModTime())
	}

	log.Println("Remote")
	remoteFiles, err := d.RemoteFiles()

	log.Println("Diff")
	uploads, removals := fileMap.Diff(remoteFiles)
	log.Println("# Uploads:")
	for filePath := range uploads {
		log.Printf("%s", filePath)
	}
	log.Println("# Removals:")
	for filePath := range removals {
		log.Printf("%s", filePath)
	}

	return nil
}

func (d *Deployment) RemoteFiles() (walk.FileMap, error) {
	err := d.sftpClient.Connect()
	if err != nil {
		return nil, err
	}

	files, errors := d.sftpClient.RemoteFiles()
	if errors != nil {
		var errorsString string
		for _, err := range errors {
			errorsString += err.Error() + "\n"
		}

		return nil, fmt.Errorf("errors getting remote files:%s\n", errorsString)
	}

	for filePath, file := range files {
		log.Printf("%s: name - %s (is dir? %t), lastMod - %s", filePath, file.Name(), file.IsDir(), file.ModTime())
	}

	_ = d.sftpClient.Close()

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
