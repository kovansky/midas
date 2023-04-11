/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package http

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog"
	"github.com/kovansky/midas"
	"github.com/kovansky/midas/strapi"
	"github.com/rs/zerolog"
	"io"
	"net/http"
)

type StrapiToAstroHandler struct {
	AstroSite midas.SiteService
	Payload   midas.Payload
	log       zerolog.Logger
}

func (s *Server) registerStrapiToAstroRoutes(r chi.Router) {
	r.Post("/strapi/astro", s.handleStrapiToAstro)
	r.Post("/strapi/astro/rebuild", s.HandleAstroRebuild)
}

func (s *Server) handleStrapiToAstro(w http.ResponseWriter, r *http.Request) {
	log := httplog.LogEntry(r.Context())

	log.Info().Msgf("Received request from strapi to astro")

	cfg := midas.SiteConfigFromContext(r.Context())

	if cfg == nil {
		Error(w, r, midas.Errorf(midas.ErrInternal, "site config not passed to the handler"))
		return
	}

	if cfg.Service != "astro" {
		Error(w, r, midas.Errorf(midas.ErrInvalid, "service mismatch: called %s while requested site is on %s", "astro", cfg.Service))
		return
	}

	jsonBody, err := io.ReadAll(r.Body)

	if err != nil {
		Error(w, r, err)
		return
	}

	payload, err := strapi.ParsePayload(jsonBody)
	var syntxErr *json.SyntaxError

	if err != nil && errors.As(err, &syntxErr) {
		Error(w, r, midas.Errorf(midas.ErrInvalid, "payload JSON malformed"))
		return
	} else if err != nil {
		Error(w, r, err)
		return
	}

	log.Info().Fields(map[string]interface{}{
		"page":  cfg.SiteName,
		"model": payload.Metadata()["model"],
		"id":    payload.Entry()["id"],
	}).Msg("Request data")

	astroSite, err := s.SiteServices["astro"](*cfg)
	if err != nil {
		Error(w, r, err)
		return
	}

	handler := &StrapiToAstroHandler{
		AstroSite: astroSite,
		Payload:   payload,
		log:       log,
	}
	defer func() {
		registry, _ := handler.AstroSite.GetRegistryService()
		registry.CloseStorage()
	}()

	handler.Handle(w, r)
}

func (s *Server) HandleAstroRebuild(w http.ResponseWriter, r *http.Request) {
	log := httplog.LogEntry(r.Context())

	log.Info().Msgf("Received rebuild request from strapi to astro")

	cfg := midas.SiteConfigFromContext(r.Context())

	if cfg == nil {
		Error(w, r, midas.Errorf(midas.ErrInternal, "site config not passed to the handler"))
		return
	}

	if cfg.Service != "astro" {
		Error(w, r, midas.Errorf(midas.ErrInvalid, "service mismatch: called %s while requested site is on %s", "astro", cfg.Service))
		return
	}

	astroSite, err := s.SiteServices["astro"](*cfg)
	if err != nil {
		Error(w, r, err)
		return
	}

	handler := &StrapiToAstroHandler{
		AstroSite: astroSite,
		Payload:   &strapi.Payload{},
		log:       log,
	}

	useCache := true
	if r.URL.Query().Has("cache") {
		switch r.URL.Query().Get("cache") {
		case "0", "false", "disable":
			useCache = false
		}
	}

	if err = astroSite.BuildSite(useCache, log); err != nil {
		Error(w, r, err)
		return
	}

	if err := handler.runDeploys(r); err != nil {
		Error(w, r, err)
		return
	}

	response := map[string]string{
		"status": "ok",
	}

	jsoned, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsoned)
}

func (h StrapiToAstroHandler) Handle(w http.ResponseWriter, r *http.Request) {
	cfg := midas.SiteConfigFromContext(r.Context())

	model := h.Payload.Metadata()["model"].(string)
	_, isSingle := cfg.SingleTypes[model]
	_, isCollection := cfg.CollectionTypes[model]

	if !isSingle && !isCollection {
		Error(w, r, midas.Errorf(midas.ErrUnaccepted, "model %s is not accepted", model))
		return
	}

	switch h.Payload.Event() {
	case strapi.Create.String(), strapi.Update.String(), strapi.Delete.String():
		h.handleBuild(w, r)
		return
	default:
		Error(w, r, midas.Errorf(midas.ErrInvalid, "event %s is invalid", h.Payload.Event()))
		return
	}
}

func (h StrapiToAstroHandler) handleBuild(w http.ResponseWriter, r *http.Request) {
	if err := h.AstroSite.BuildSite(true, h.log); err != nil {
		Error(w, r, err)
		return
	}

	if err := h.runDeploys(r); err != nil {
		Error(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// runDeploys executes both final and the draft deploys.
func (h StrapiToAstroHandler) runDeploys(r *http.Request) error {
	cfg := midas.SiteConfigFromContext(r.Context())

	if err := h.deploy(cfg, false); err != nil {
		return err
	}

	if err := h.deploy(cfg, true); err != nil {
		return err
	}

	return nil
}

// deploy executes the uploading process.
func (h StrapiToAstroHandler) deploy(cfg *midas.Site, draft bool) error {
	var dplSettings midas.DeploymentSettings
	if draft {
		dplSettings = cfg.DraftsDeployment
	} else {
		dplSettings = cfg.Deployment
	}

	if !dplSettings.Enabled {
		return nil
	}

	var deploymentService midas.Deployment
	if dpl, ok := midas.DeploymentTargets[dplSettings.Target]; ok {
		var err error

		if deploymentService, err = dpl(*cfg, dplSettings, draft); err != nil {
			return midas.Errorf(midas.ErrInternal, "could not create deployment %s: %s", dplSettings.Target, err)
		}
	} else {
		return midas.Errorf(midas.ErrUnaccepted, "deployment target %s is not accepted", dplSettings.Target)
	}

	h.log.Debug().Msgf("Deploying %s to %s", cfg.SiteName, dplSettings.Target)
	if err := deploymentService.Deploy(); err != nil {
		return err
	}

	return nil
}
