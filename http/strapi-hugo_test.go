package http_test

import (
	"bytes"
	"context"
	"github.com/kovansky/midas"
	"github.com/kovansky/midas/mock"
	"net/http"
	"testing"
)

var (
	BuildSiteCounter   = 0
	CreateEntryCounter = 0
)

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

			var testTable = map[string][]interface{}{
				"StatusCode":       {resp.StatusCode, http.StatusNoContent},
				"BuildSiteCounter": {BuildSiteCounter, 1},
			}

			for name, values := range testTable {
				got, want := values[0], values[1]

				if got != want {
					t.Fatalf("%s=%v, want %v", name, got, want)
				}
			}
		})
	})
}
