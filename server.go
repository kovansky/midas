package strapi2hugo

import (
	"encoding/json"
	"fmt"
	"github.com/kovansky/strapi2hugo/models"
	"github.com/kovansky/strapi2hugo/models/enums"
	"github.com/kovansky/strapi2hugo/utils"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

var config models.Config

func main() {
	fmt.Println("Reading config")
	config = readConfig()

	fmt.Printf("Listening on :%d\n", config.ListeningPort)
	http.HandleFunc("/", buildHandler)
	fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", config.ListeningPort), nil))
}

func readConfig() models.Config {
	configFile, err := os.Open("config.json")
	defer func(configFile *os.File) {
		_ = configFile.Close()
	}(configFile)

	if err != nil {
		panic(err)
	}

	fileBody, err := ioutil.ReadAll(configFile)

	if err != nil {
		panic(err)
	}

	var config models.Config

	err = json.Unmarshal(fileBody, &config)

	return config
}

func buildHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Request at %v\n", time.Now())

	if r.Method == "POST" {
		// Check, if api key is set
		if r.Header.Get("Authorization") != "" {
			// Get api key
			apiKey := strings.Replace(r.Header.Get("Authorization"), "Bearer ", "", -1)

			// Check, if api key is correct and if there is site registered for this key
			if site, ok := config.Sites[apiKey]; ok {
				// Read payload
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					fmt.Println(err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				var payload models.WebhookPayload
				err = json.Unmarshal(body, &payload)
				if err != nil {
					fmt.Println(err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				if payload.Event == enums.Create {
					if utils.Contains(site.CollectionTypes, payload.Model) {
						fmt.Printf("Received %s event for %s site\n", payload.Event, site.SiteName)

						// Try to write new file
						if ok = site.CreateEntry(payload); !ok {
							fmt.Println("Could not create entry")
							w.WriteHeader(http.StatusInternalServerError)
							return
						}

						// Rebuild site without ignoring cache
						if ok = site.RebuildSite(false); !ok {
							fmt.Println("Building site error")
							w.WriteHeader(http.StatusInternalServerError)
							return
						}

						fmt.Println("Added entry and rebuilt website")
						w.WriteHeader(http.StatusOK)
						return
					} else {
						fmt.Printf("Event %s not implemented for model %s\n", payload.Event, payload.Model)
						w.WriteHeader(http.StatusNotImplemented)
						return
					}
				} else if payload.Event == enums.Update {
					if utils.Contains(site.SingleTypes, payload.Model) {
						// Rebuild site with ignoring cache
						if ok = site.RebuildSite(true); !ok {
							fmt.Println("Building site error")
							w.WriteHeader(http.StatusInternalServerError)
							return
						}

						fmt.Println("Rebuilt website")
						w.WriteHeader(http.StatusOK)
						return
					} else {
						fmt.Printf("Event %s not implemented for model %s\n", payload.Event, payload.Model)
						w.WriteHeader(http.StatusNotImplemented)
						return
					}
				} else {
					fmt.Printf("Event %s not implemented\n", payload.Event)
					w.WriteHeader(http.StatusNotImplemented)
					return
				}
			} else {
				fmt.Println("Given API key not recognized")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		} else {
			fmt.Println("No API key provided")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	} else {
		fmt.Printf("Method %s not implemented\n", r.Method)
		w.WriteHeader(http.StatusNotImplemented)
	}
}
