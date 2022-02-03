package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/kovansky/midas"
	midashttp "github.com/kovansky/midas/http"
	"github.com/kovansky/midas/mock"
	"github.com/kovansky/midas/testing_utils"
	"io"
	"net/http"
	"testing"
)

var (
	BuildSiteCounter   = 0
	CreateEntryCounter = 0
)

func resetCounters() {
	BuildSiteCounter = 0
	CreateEntryCounter = 0
}

func SetUp(t *testing.T) *Server {
	s := MustOpenServer(t, map[string]func(site midas.Site) midas.SiteService{
		"hugo": func(_ midas.Site) midas.SiteService {
			siteService := mock.NewSiteService()

			siteService.BuildSiteFn = func(useCache bool) error {
				BuildSiteCounter++

				return nil
			}

			siteService.CreateEntryFn = func(_ midas.Payload) (string, error) {
				CreateEntryCounter++

				return "", nil
			}

			return siteService
		},
	}, midas.Config{
		Sites: map[string]midas.Site{
			"test": {
				CollectionTypes: []string{"post"},
				SingleTypes:     []string{"homepage"},
			},
		},
	})

	return s
}

func TestStrapiToHugoHandler_Handle(t *testing.T) {
	s := SetUp(t)
	defer MustCloseServer(t, s)

	t.Run("Unauthenticated", func(t *testing.T) {
		t.Run("NoKey", func(t *testing.T) {
			resp, err := http.DefaultClient.Do(s.MustNewRequest(t, context.Background(), "", "POST", "/strapi/hugo", bytes.NewReader([]byte(""))))
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
			resp, err := http.DefaultClient.Do(s.MustNewRequest(t, context.Background(), "xyz", "POST", "/strapi/hugo", bytes.NewReader([]byte(""))))
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

			resp, err := http.DefaultClient.Do(s.MustNewRequest(t, context.Background(), "test", "POST", "/strapi/hugo", bytes.NewReader(jsonPayload)))
			if err != nil {
				t.Fatal(err)
			}

			testing_utils.AssertTable(t, map[string][]interface{}{
				"Status code":      {resp.StatusCode, http.StatusNoContent},
				"BuildSiteCounter": {BuildSiteCounter, 1},
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

			resp, err := http.DefaultClient.Do(s.MustNewRequest(t, context.Background(), "test", "POST", "/strapi/hugo", bytes.NewReader(jsonPayload)))
			if err != nil {
				t.Fatal(err)
			}

			testing_utils.AssertTable(t, map[string][]interface{}{
				"Status code":        {resp.StatusCode, http.StatusNoContent},
				"CreateEntryCounter": {CreateEntryCounter, 1},
				"BuildSiteCounter":   {BuildSiteCounter, 1},
			})
		})

		resetCounters()

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

			resp, err := http.DefaultClient.Do(s.MustNewRequest(t, context.Background(), "test", "POST", "/strapi/hugo", bytes.NewReader(jsonPayload)))
			if err != nil {
				t.Fatal(err)
			}

			testing_utils.AssertTable(t, map[string][]interface{}{
				"Status code": {resp.StatusCode, http.StatusBadRequest},
			})
		})
	})
}
