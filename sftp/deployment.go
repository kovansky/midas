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
		publicPath:         filepath.ToSlash(publicPath),

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
	diff := fileMap.Diff(remoteFiles)

	err = d.sftpClient.Connect()
	if err != nil {
		return err
	}
	defer func(sftpClient *Client) {
		_ = sftpClient.Close()
	}(&d.sftpClient)

	for _, fileOp := range diff {
		err := d.syncFile(fileOp)
		if err != nil {
			return err
		}
	}

	return nil
}

// syncFile performs a file operation.
func (d *Deployment) syncFile(operation walk.FileOperation) error {
	switch operation.Type {
	case walk.UploadFile, walk.UpdateFile:
		absolute := filepath.ToSlash(filepath.Clean(filepath.Join(d.publicPath, operation.Path)))

		handler, err := os.Open(absolute)
		if err != nil {
			return err
		}
		defer func(handler *os.File) {
			_ = handler.Close()
		}(handler)

		if err = d.sftpClient.UploadNewFile(operation.Path, handler); err != nil {
			return err
		}

		break
	case walk.RemoveFile:
		if err := d.sftpClient.RemoveFile(operation.Path); err != nil {
			return err
		}
	}

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
		relPath = filepath.ToSlash(relPath)

		fileMap[relPath] = fileInfo
	}

	return fileMap, nil
}
