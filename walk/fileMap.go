/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package walk

import "os"

// FileOperationType is used to identify what kind of operation should be performed on a file.
type FileOperationType int64

const (
	// UploadFile means that file should be uploaded (created).
	UploadFile FileOperationType = iota
	// UpdateFile means that file should be uploaded (modified).
	UpdateFile
	// RemoveFile means that file should be removed.
	RemoveFile
)

// FileMap is type for holding a map of files information indexed by their name (relative path).
type FileMap map[string]os.FileInfo

type FileOperation struct {
	Path string
	Info os.FileInfo
	Type FileOperationType
}

// Diff compares the two filetrees and returns a list of differences (files to be removed, updated or created).
//
// The differences are relative to the calling FileMap, which means that for example UploadFile opearion should transfer
// the file FROM the calling FileMap location to the other FileMap location.
func (f FileMap) Diff(other FileMap) (diff []FileOperation) {
	for name, info := range f {
		if otherInfo, ok := other[name]; !ok {
			diff = append(diff, FileOperation{
				Path: name,
				Info: info,
				Type: UploadFile,
			})
		} else {
			if info.ModTime().After(otherInfo.ModTime()) {
				diff = append(diff, FileOperation{
					Path: name,
					Info: info,
					Type: UpdateFile,
				})
			}
		}
	}

	for name, info := range other {
		if _, ok := f[name]; !ok {
			diff = append(diff, FileOperation{
				Path: name,
				Info: info,
				Type: RemoveFile,
			})
		}
	}

	return
}
