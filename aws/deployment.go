/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kovansky/midas"
	"os"
	"path/filepath"
	"strings"
)

type fileWalk chan string

func (f fileWalk) Walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	f <- path
	return nil
}

type Deployment struct {
	site       midas.Site
	publicPath string
}

func New(site midas.Site) midas.Deployment {
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

	return &Deployment{site: site, publicPath: publicPath}
}

// Deploy uploads built site to the AWS S3 bucket.
func (d *Deployment) Deploy() error {
	walker, err := d.retrieveFiles()
	if err != nil {
		return err
	}

	accessKey, secretKey := d.retrieveKeys()

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")))
	if err != nil {
		return err
	}

	// Upload each file to the S3 bucket.
	uploader := manager.NewUploader(s3.NewFromConfig(cfg))
	for path := range walker {
		err = func() error {
			rel, err := filepath.Rel(d.publicPath, path)
			if err != nil {
				return err
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer func() {
				_ = file.Close()
			}()

			if err = d.uploadFile(uploader, file, rel); err != nil {
				return err
			}

			return nil
		}()
		if err != nil {
			return err
		}
	}

	return nil
}

// uploadFile uploads a file to the S3 bucket.
func (d *Deployment) uploadFile(uploader *manager.Uploader, file *os.File, rel string) error {
	fileKey := rel
	if d.site.Deployment.Additional["prefix"] != "" {
		fileKey = filepath.Join(d.site.Deployment.Additional["prefix"], rel)
	}

	_, err := uploader.Upload(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(d.site.Deployment.Additional["bucket"]),
		Key:    aws.String(fileKey),
		Body:   file,
	})
	if err != nil {
		return err
	}

	return nil
}

// reteiveFiles walks the public directory and returns a channel of files to be uploaded.
func (d *Deployment) retrieveFiles() (fileWalk, error) {
	walker := make(fileWalk)
	routineErr := make(chan error)

	// Gather the files to upload by walking the path recursively.
	go func() {
		defer close(walker)
		if err := filepath.Walk(d.publicPath, walker.Walk); err != nil {
			routineErr <- err
		}
	}()

	if err := <-routineErr; err != nil {
		return nil, err
	}

	return walker, nil
}

// retrieveKeys splits the key provided in config to access key and secret key in given order.
func (d Deployment) retrieveKeys() (string, string) {
	sliced := strings.Split(d.site.Deployment.Key, "|")
	return sliced[0], sliced[1]
}
