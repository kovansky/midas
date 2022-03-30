/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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
	site               midas.Site
	deploymentSettings midas.DeploymentSettings
	publicPath         string
	s3Client           *s3.Client
}

func New(site midas.Site, deploymentSettings midas.DeploymentSettings) midas.Deployment {
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

	return &Deployment{site: site, deploymentSettings: deploymentSettings, publicPath: publicPath}
}

// Deploy uploads built site to the AWS S3 bucket.
func (d *Deployment) Deploy() error {
	walker, err := d.retrieveFiles()
	if err != nil {
		return err
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(d.deploymentSettings.AWS.AccessKey, d.deploymentSettings.AWS.SecretKey, "")),
		config.WithRegion(d.deploymentSettings.AWS.Region))
	if err != nil {
		return err
	}

	d.s3Client = s3.NewFromConfig(cfg)

	var currentObjects []string
	if currentObjects, err = d.listObjects(); err != nil {
		return err
	}

	_ = d.deleteObjects(currentObjects)

	// Upload each file to the S3 bucket.
	uploader := manager.NewUploader(d.s3Client)
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

// ToDo: Cloudfront invalidation

// uploadFile uploads a file to the S3 bucket.
func (d *Deployment) uploadFile(uploader *manager.Uploader, file *os.File, rel string) error {
	fileKey := rel
	if d.deploymentSettings.AWS.S3Prefix != "" {
		fileKey = fmt.Sprintf("%s/%s", d.deploymentSettings.AWS.S3Prefix, rel)
	}

	fileKey = strings.ReplaceAll(fileKey, "\\", "/")

	contentType := getFileContentType(file.Name())

	_, err := uploader.Upload(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(d.deploymentSettings.AWS.BucketName),
		Key:         aws.String(fileKey),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return err
	}

	return nil
}

// listObjects retrieves a list of objects in the S3 bucket.
func (d *Deployment) listObjects() ([]string, error) {
	var objects []string

	output, err := d.s3Client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket: aws.String(d.deploymentSettings.AWS.BucketName),
	})
	if err != nil {
		return nil, err
	}

	for _, obj := range output.Contents {
		objects = append(objects, aws.ToString(obj.Key))
	}

	return objects, nil
}

// deleteObjects deletes objects from the S3 bucket.
func (d *Deployment) deleteObjects(objects []string) error {
	var identifiers []types.ObjectIdentifier
	for key := range objects {
		identifiers = append(identifiers, types.ObjectIdentifier{
			Key: aws.String(objects[key]),
		})
	}

	_, err := d.s3Client.DeleteObjects(context.Background(), &s3.DeleteObjectsInput{
		Bucket: aws.String(d.deploymentSettings.AWS.BucketName),
		Delete: &types.Delete{
			Objects: identifiers,
		},
	})

	if err != nil {
		return err
	}

	return nil
}

// reteiveFiles walks the public directory and returns a channel of files to be uploaded.
func (d *Deployment) retrieveFiles() (fileWalk, error) {
	walker := make(fileWalk)

	// Gather the files to upload by walking the path recursively.
	go func() {
		defer close(walker)
		if err := filepath.Walk(d.publicPath, walker.Walk); err != nil {
			panic(err)
		}
	}()

	return walker, nil
}

// getFileContentType returns the content type of the file based on the extension.
func getFileContentType(fileName string) string {
	typeByExtension := map[string]string{
		".html": "text/html",
		".css":  "text/css",
		".xml":  "text/xml",

		".js":  "application/javascript",
		".pdf": "application/pdf",

		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".svg":  "image/svg+xml",
		".webp": "image/webp",

		".webm": "video/webm",
		".mp4":  "video/mp4",
		".ogv":  "video/ogg",
		".avi":  "video/x-msvideo",

		".ogg":  "audio/ogg",
		".mp3":  "audio/mpeg",
		".mpeg": "audio/mpeg",
	}

	extension := filepath.Ext(fileName)

	if contentType, ok := typeByExtension[extension]; ok {
		return contentType
	} else {
		return "application/octet-stream"
	}
}
