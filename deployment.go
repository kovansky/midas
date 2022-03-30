/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package midas

type Deployment interface {
	Deploy() error
}

type DeploymentSettings struct {
	Enabled bool                 `json:"enabled,default=false"`
	Target  string               `json:"target"` // Can be: AWS
	AWS     AWSDeploymentSettigs `json:"aws,omitempty"`
}

type AWSDeploymentSettigs struct {
	BucketName string `json:"bucketName"`
	Region     string `json:"region"`
	AccessKey  string `json:"accessKey"`
	SecretKey  string `json:"secretKey"`
	S3Prefix   string `json:"s3Prefix"`
}
