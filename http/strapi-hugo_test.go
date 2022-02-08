package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	midashttp "github.com/kovansky/midas/http"
	"github.com/kovansky/midas/testing_utils"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestStrapiToHugoHandler_Handle(t *testing.T) {
	endpoint := "/strapi/hugo"

	s := SetUp(t)
	defer MustCloseServer(t, s)

	t.Run("Unauthenticated", func(t *testing.T) {
		t.Run("NoKey", func(t *testing.T) {
			resp, err := http.DefaultClient.Do(s.MustNewRequest(t, context.Background(), "", "POST", endpoint, bytes.NewReader([]byte(""))))
			if err != nil {
				t.Fatal(err)
			}

			testing_utils.AssertEquals(t, resp.StatusCode, http.StatusUnauthorized, "Status code")

			jsonBody, _ := io.ReadAll(resp.Body)
			var respError midashttp.ErrorResponse
			err = json.Unmarshal(jsonBody, &respError)

			testing_utils.AssertTable(t, map[string][]interface{}{
				"Response body json unmarshal error": {err, nil},
				"Error response":                     {respError, midashttp.ErrorResponse{Error: "No API key."}},
			})
		})
		t.Run("IncorrectKey", func(t *testing.T) {
			resp, err := http.DefaultClient.Do(s.MustNewRequest(t, context.Background(), "xyz", "POST", endpoint, bytes.NewReader([]byte(""))))
			if err != nil {
				t.Fatal(err)
			}

			testing_utils.AssertEquals(t, resp.StatusCode, http.StatusUnauthorized, "Status code")

			jsonBody, _ := io.ReadAll(resp.Body)
			var respError midashttp.ErrorResponse
			err = json.Unmarshal(jsonBody, &respError)

			testing_utils.AssertTable(t, map[string][]interface{}{
				"Response body json unmarshal error": {err, nil},
				"Error response":                     {respError, midashttp.ErrorResponse{Error: "Invalid API key."}},
			})
		})
	})

	t.Run("ServiceMismatch", func(t *testing.T) {
		resp, err := http.DefaultClient.Do(s.MustNewRequest(t, context.Background(), "otherService", "POST", endpoint, bytes.NewReader([]byte(""))))
		if err != nil {
			t.Fatal(err)
		}

		testing_utils.AssertEquals(t, resp.StatusCode, http.StatusBadRequest, "Status code")

		jsonBody, _ := io.ReadAll(resp.Body)
		var respError midashttp.ErrorResponse
		err = json.Unmarshal(jsonBody, &respError)

		testing_utils.AssertTable(t, map[string][]interface{}{
			"Response body json unmarshal error": {err, nil},
			"Error response":                     {strings.HasPrefix(respError.Error, "service mismatch:"), true},
		})
	})

	t.Run("UnsupportedModel", func(t *testing.T) {
		jsonPayload := []byte(`{
    "event": "entry.create",
    "createdAt": "2022-01-01T10:10:10.000Z",
    "model": "review",
    "entry": {
      "id": 1,
      "Title": "Test",
      "Content": "Test",
      "createdAt": "2022-01-01T10:10:10.000Z",
      "updatedAt": "2022-01-01T10:10:10.000Z",
      "publishedAt": null
    }
  }`)

		resp, err := http.DefaultClient.Do(s.MustNewRequest(t, context.Background(), "test", "POST", endpoint, bytes.NewReader(jsonPayload)))
		if err != nil {
			t.Fatal(err)
		}

		testing_utils.AssertTable(t, map[string][]interface{}{
			"Status code": {resp.StatusCode, http.StatusBadRequest},
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("Single", func(t *testing.T) {
			jsonPayload := []byte(`{
    "event": "entry.create",
    "createdAt": "2022-01-01T10:10:10.000Z",
    "model": "homepage",
    "entry": {
      "id": 1,
      "Title": "Test",
      "Content": "Test",
      "createdAt": "2022-01-01T10:10:10.000Z",
      "updatedAt": "2022-01-01T10:10:10.000Z",
      "publishedAt": null
    }
  }`)

			resp, err := http.DefaultClient.Do(s.MustNewRequest(t, context.Background(), "test", "POST", endpoint, bytes.NewReader(jsonPayload)))
			if err != nil {
				t.Fatal(err)
			}

			testing_utils.AssertTable(t, map[string][]interface{}{
				"Status code":    {resp.StatusCode, http.StatusNoContent},
				"Site.BuildSite": {MockSiteCounters["BuildSite"], 1},
			})
		})

		resetCounters()

		t.Run("Collection", func(t *testing.T) {
			jsonPayload := []byte(`{
    "event": "entry.create",
    "createdAt": "2022-01-01T10:10:10.000Z",
    "model": "post",
    "entry": {
      "id": 1,
      "Title": "Test",
      "Content": "Test",
      "createdAt": "2022-01-01T10:10:10.000Z",
      "updatedAt": "2022-01-01T10:10:10.000Z",
      "publishedAt": null
    }
  }`)

			resp, err := http.DefaultClient.Do(s.MustNewRequest(t, context.Background(), "test", "POST", endpoint, bytes.NewReader(jsonPayload)))
			if err != nil {
				t.Fatal(err)
			}

			testing_utils.AssertTable(t, map[string][]interface{}{
				"Status code":      {resp.StatusCode, http.StatusNoContent},
				"Site.CreateEntry": {MockSiteCounters["CreateEntry"], 1},
				"Site.BuildSite":   {MockSiteCounters["BuildSite"], 1},
			})
		})
	})

	resetCounters()

	t.Run("Update", func(t *testing.T) {
		t.Run("Single", func(t *testing.T) {
			jsonPayload := []byte(`{
    "event": "entry.update",
    "createdAt": "2022-01-01T10:10:10.000Z",
    "model": "homepage",
    "entry": {
      "id": 1,
      "Title": "Test",
      "Content": "Test",
      "createdAt": "2022-01-01T10:10:10.000Z",
      "updatedAt": "2022-01-01T10:10:10.000Z",
      "publishedAt": null
    }
  }`)

			resp, err := http.DefaultClient.Do(s.MustNewRequest(t, context.Background(), "test", "POST", endpoint, bytes.NewReader(jsonPayload)))
			if err != nil {
				t.Fatal(err)
			}

			testing_utils.AssertTable(t, map[string][]interface{}{
				"Status code":    {resp.StatusCode, http.StatusNoContent},
				"Site.BuildSite": {MockSiteCounters["BuildSite"], 1},
			})
		})

		resetCounters()

		t.Run("Collection", func(t *testing.T) {
			jsonPayload := []byte(`{
    "event": "entry.update",
    "createdAt": "2022-01-01T10:10:10.000Z",
    "model": "post",
    "entry": {
      "id": 1,
      "Title": "Test",
      "Content": "Test",
      "createdAt": "2022-01-01T10:10:10.000Z",
      "updatedAt": "2022-01-01T10:10:10.000Z",
      "publishedAt": null
    }
  }`)

			resp, err := http.DefaultClient.Do(s.MustNewRequest(t, context.Background(), "test", "POST", endpoint, bytes.NewReader(jsonPayload)))
			if err != nil {
				t.Fatal(err)
			}

			testing_utils.AssertTable(t, map[string][]interface{}{
				"Status code":      {resp.StatusCode, http.StatusNoContent},
				"Site.UpdateEntry": {MockSiteCounters["UpdateEntry"], 1},
				"Site.BuildSite":   {MockSiteCounters["BuildSite"], 1},
			})
		})
	})
}
