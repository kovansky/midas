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
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/kovansky/midas"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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

var _ midas.Deployment = (*Deployment)(nil)

type Deployment struct {
	site               midas.Site
	deploymentSettings midas.DeploymentSettings
	publicPath         string

	awsConfig aws.Config
	s3Client  *s3.Client
	cfClient  *cloudfront.Client
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

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(deploymentSettings.AWS.AccessKey, deploymentSettings.AWS.SecretKey, "")),
		config.WithRegion(deploymentSettings.AWS.Region))
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(cfg)
	cfClient := cloudfront.NewFromConfig(cfg)

	return &Deployment{site: site, deploymentSettings: deploymentSettings, publicPath: publicPath, s3Client: s3Client, cfClient: cfClient}, nil
}

// Deploy uploads built site to the AWS S3 bucket.
func (d *Deployment) Deploy() error {
	walker, err := d.retrieveFiles()
	if err != nil {
		return err
	}

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

	err = d.invalidateCloudfront()
	if err != nil {
		return err
	}

	return nil
}

// uploadFile uploads a file to the S3 bucket.
func (d *Deployment) uploadFile(uploader *manager.Uploader, file *os.File, rel string) error {
	fileKey := rel
	if d.deploymentSettings.AWS.S3Prefix != "" {
		fileKey = fmt.Sprintf("%s/%s", d.deploymentSettings.AWS.S3Prefix, rel)
	}

	fileKey = strings.ReplaceAll(fileKey, "\\", "/")

	contentType := getFileContentType(file.Name())
	cacheControl := getFileCacheControl(file.Name())

	_, err := uploader.Upload(context.Background(), &s3.PutObjectInput{
		Bucket:       aws.String(d.deploymentSettings.AWS.BucketName),
		Key:          aws.String(fileKey),
		Body:         file,
		ContentType:  aws.String(contentType),
		CacheControl: aws.String(cacheControl),
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
	var identifiers []s3types.ObjectIdentifier
	for key := range objects {
		identifiers = append(identifiers, s3types.ObjectIdentifier{
			Key: aws.String(objects[key]),
		})
	}

	_, err := d.s3Client.DeleteObjects(context.Background(), &s3.DeleteObjectsInput{
		Bucket: aws.String(d.deploymentSettings.AWS.BucketName),
		Delete: &s3types.Delete{
			Objects: identifiers,
		},
	})

	if err != nil {
		return err
	}

	return nil
}

// invalidateCloudfront invalidates the HTML files in the Cloudfront distribution.
func (d *Deployment) invalidateCloudfront() error {
	paths := []string{"/*"}

	if d.deploymentSettings.AWS.CloudfrontDistribution != "" {
		_, err := d.cfClient.CreateInvalidation(context.Background(), &cloudfront.CreateInvalidationInput{
			DistributionId: aws.String(d.deploymentSettings.AWS.CloudfrontDistribution),
			InvalidationBatch: &cftypes.InvalidationBatch{
				CallerReference: aws.String(fmt.Sprintf("MIDAS-%s", strconv.FormatInt(time.Now().Unix(), 10))),
				Paths: &cftypes.Paths{
					Quantity: aws.Int32(int32(len(paths))),
					Items:    paths,
				},
			},
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// retrieveFiles walks the public directory and returns a channel of files to be uploaded.
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

// getFileCacheControl returns the max-age value for the file based on it's type.
func getFileCacheControl(fileName string) string {
	halfYear := int64(60 * 60 * 24 * 182)

	switch {
	case strings.HasSuffix(fileName, ".html"):
		return "no-cache, no-store"
	case strings.HasSuffix(fileName, ".js"), strings.HasSuffix(fileName, ".css"),
		strings.HasSuffix(fileName, ".svg"), strings.HasSuffix(fileName, ".png"),
		strings.HasSuffix(fileName, ".jpg"), strings.HasSuffix(fileName, ".jpeg"),
		strings.HasSuffix(fileName, ".gif"):
		return fmt.Sprintf("public, max-age=%d", halfYear)
	default:
		return fmt.Sprintf("public, max-age=%d", halfYear)
	}
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
