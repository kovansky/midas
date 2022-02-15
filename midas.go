/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package midas

import (
	"context"
	"log"
)

var (
	Commit  string
	Version string

	RegistryServices map[string]func(site Site) RegistryService
	Sanitizer        SanitizerService
)

// ReportError is used to notify external services of error.
var ReportError = func(ctx context.Context, err error, args ...interface{}) {
	log.Printf("error: %+v\n", err)
}

// ReportPanic is used to notify external services of panic. Maybe will be used in future.
var _ = func(err interface{}) {
	log.Panic(err)
}
