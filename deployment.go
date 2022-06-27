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
	Enabled bool                   `json:"enabled,default=false"`
	Target  string                 `json:"target"` // Can be: AWS, SFTP
	AWS     AWSDeploymentSettigs   `json:"aws,omitempty"`
	SFTP    SFTPDeploymentSettings `json:"sftp,omitempty"`
}

type AWSDeploymentSettigs struct {
	BucketName             string `json:"bucketName"`
	Region                 string `json:"region"`
	AccessKey              string `json:"accessKey"`
	SecretKey              string `json:"secretKey"`
	S3Prefix               string `json:"s3Prefix"`
	CloudfrontDistribution string `json:"cloudfrontDistribution,omitempty"`
}

type SFTPDeploymentSettings struct {
	Host          string `json:"host"`
	Port          *int   `json:"port"`
	User          string `json:"user"`
	Method        string `json:"method"`
	Password      string `json:"password,omitempty"`
	Key           string `json:"key,omitempty"`
	KeyPassphrase string `json:"keyPassphrase,omitempty"`
	Path          string `json:"path"`
}
