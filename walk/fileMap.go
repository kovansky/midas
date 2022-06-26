/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package walk

import "os"

// FileMap is type for holding a map of files information indexed by their name (relative path).
type FileMap map[string]os.FileInfo