/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package midas

import "github.com/kovansky/midas/aws"

type Deployment interface {
	Deploy() error
}

type DeploymentSettings struct {
	Enabled bool                  `json:"enabled,default=false"`
	Target  string                `json:"target"` // Can be: AWS
	AWS     aws.DeploymentSettigs `json:"aws,omitempty"`
}
