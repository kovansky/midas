/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/kovansky/midas"
	midashttp "github.com/kovansky/midas/http"
	"github.com/kovansky/midas/testing_utils"
	"io"
	"net/http"
	"testing"
)

func TestServer_HandleSystemCommit(t *testing.T) {
	endpoint := "/system/commit"

	s := SetUp(t)
	defer MustCloseServer(t, s)

	resp, err := http.DefaultClient.Do(s.MustNewRequest(t, context.Background(), "", "GET", endpoint, bytes.NewReader([]byte(""))))
	if err != nil {
		t.Fatal(err)
	}

	testing_utils.AssertEquals(t, resp.StatusCode, http.StatusOK, "Status code")

	jsonBody, _ := io.ReadAll(resp.Body)
	var parsedBody midashttp.SystemResponse
	err = json.Unmarshal(jsonBody, &parsedBody)

	testing_utils.AssertTable(t, map[string][]interface{}{
		"Response body json unmarshal error": {err, nil},
		"System response":                    {parsedBody, midashttp.SystemResponse{Data: midas.Commit}},
	})
}

func TestServer_HandleSystemVersion(t *testing.T) {
	endpoint := "/system/version"

	s := SetUp(t)
	defer MustCloseServer(t, s)

	resp, err := http.DefaultClient.Do(s.MustNewRequest(t, context.Background(), "", "GET", endpoint, bytes.NewReader([]byte(""))))
	if err != nil {
		t.Fatal(err)
	}

	testing_utils.AssertEquals(t, resp.StatusCode, http.StatusOK, "Status code")

	jsonBody, _ := io.ReadAll(resp.Body)
	var parsedBody midashttp.SystemResponse
	err = json.Unmarshal(jsonBody, &parsedBody)

	testing_utils.AssertTable(t, map[string][]interface{}{
		"Response body json unmarshal error": {err, nil},
		"System response":                    {parsedBody, midashttp.SystemResponse{Data: midas.Version}},
	})
}
