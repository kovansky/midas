/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package testing_utils

import "testing"

func AssertEquals(t *testing.T, got, want interface{}, name ...string) {
	displayName := ""

	if len(name) > 0 {
		displayName = name[0] + ": "
	}

	if got != want {
		t.Fatalf("%sgot %v, want %v", displayName, got, want)
	}
}

func AssertTable(t *testing.T, table map[string][]interface{}) {
	for name, values := range table {
		got, want := values[0], values[1]

		AssertEquals(t, got, want, name)
	}
}
