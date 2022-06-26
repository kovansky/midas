/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package walk

import "os"

// FileMap is type for holding a map of files information indexed by their name (relative path).
type FileMap map[string]os.FileInfo

// Diff generates two maps of files: one with files to be uploaded (newer or added) and one with files to delete.
func (f FileMap) Diff(other FileMap) (upload, remove FileMap) {
	upload = make(FileMap)
	remove = make(FileMap)

	for name, info := range f {
		if otherInfo, ok := other[name]; !ok {
			upload[name] = info
		} else {
			if info.ModTime().After(otherInfo.ModTime()) {
				upload[name] = info
			}
		}
	}

	for name, info := range other {
		if _, ok := f[name]; !ok {
			remove[name] = info
		}
	}

	return
}
