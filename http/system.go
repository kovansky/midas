/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package http

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/kovansky/midas"
	"net/http"
)

func (s *Server) registerSystemRoutes(r chi.Router) {
	r.Get("/commit", s.HandleSystemCommit)
	r.Get("/version", s.HandleSystemVersion)
}

func (s *Server) HandleSystemCommit(w http.ResponseWriter, r *http.Request) {
	handleSystem(w, r, midas.Commit)
}

func (s *Server) HandleSystemVersion(w http.ResponseWriter, r *http.Request) {
	handleSystem(w, r, midas.Version)
}

func handleSystem(w http.ResponseWriter, _ *http.Request, value string) {
	jsoned, _ := json.Marshal(&SystemResponse{Data: value})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsoned)
}

type SystemResponse struct {
	Data string `json:"data"`
}
