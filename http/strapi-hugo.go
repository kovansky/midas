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

type StrapiToHugoHandler struct {
	HugoSite midas.SiteService
	Payload  midas.Payload
	log      zerolog.Logger
}

func (s *Server) registerStrapiToHugoRoutes(r chi.Router) {
	r.Post("/strapi/hugo", s.handleStrapiToHugo)
	r.Post("/strapi/hugo/rebuild", s.HandleHugoRebuild)
}

func (s *Server) handleStrapiToHugo(w http.ResponseWriter, r *http.Request) {
	log := httplog.LogEntry(r.Context())

	log.Info().Msgf("Received request from strapi to hugo")

	cfg := midas.SiteConfigFromContext(r.Context())

	if cfg == nil {
		Error(w, r, midas.Errorf(midas.ErrInternal, "site config not passed to the handler"))
		return
	}

	if cfg.Service != "hugo" {
		Error(w, r, midas.Errorf(midas.ErrInvalid, "service mismatch: called %s while requested site is on %s", "hugo", cfg.Service))
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

	hugoSite, err := s.SiteServices["hugo"](*cfg)
	if err != nil {
		Error(w, r, err)
		return
	}

	handler := &StrapiToHugoHandler{
		HugoSite: hugoSite,
		Payload:  payload,
		log:      log,
	}
	defer func() {
		registry, _ := handler.HugoSite.GetRegistryService()
		registry.CloseStorage()
	}()

	handler.Handle(w, r)
}

func (s *Server) HandleHugoRebuild(w http.ResponseWriter, r *http.Request) {
	log := httplog.LogEntry(r.Context())

	log.Info().Msgf("Received rebuild request from strapi to hugo")

	cfg := midas.SiteConfigFromContext(r.Context())

	if cfg == nil {
		Error(w, r, midas.Errorf(midas.ErrInternal, "site config not passed to the handler"))
		return
	}

	if cfg.Service != "hugo" {
		Error(w, r, midas.Errorf(midas.ErrInvalid, "service mismatch: called %s while requested site is on %s", "hugo", cfg.Service))
		return
	}

	hugoSite, err := s.SiteServices["hugo"](*cfg)
	if err != nil {
		Error(w, r, err)
		return
	}

	handler := &StrapiToHugoHandler{
		HugoSite: hugoSite,
		Payload:  &strapi.Payload{},
		log:      log,
	}

	useCache := true
	if r.URL.Query().Has("cache") {
		switch r.URL.Query().Get("cache") {
		case "0", "false", "disable":
			useCache = false
		}
	}

	if err = hugoSite.BuildSite(useCache); err != nil {
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

func (h StrapiToHugoHandler) Handle(w http.ResponseWriter, r *http.Request) {
	cfg := midas.SiteConfigFromContext(r.Context())

	model := h.Payload.Metadata()["model"].(string)
	var isSingle bool

	if _, ok := cfg.SingleTypes[model]; ok {
		isSingle = true
	} else if _, ok := cfg.CollectionTypes[model]; ok {
		isSingle = false
	} else {
		Error(w, r, midas.Errorf(midas.ErrUnaccepted, "model %s is not accepted", model))
		return
	}

	switch h.Payload.Event() {
	case strapi.Create.String():
		if isSingle {
			h.handleCreateSingle(w, r)
			return
		} else {
			h.handleCreateCollection(w, r)
			return
		}
	case strapi.Update.String():
		if isSingle {
			h.handleUpdateSingle(w, r)
			return
		} else {
			h.handleUpdateCollection(w, r)
			return
		}
	case strapi.Delete.String():
		if isSingle {
			break
		} else {
			h.handleDeleteCollection(w, r)
			return
		}
	default:
		Error(w, r, midas.Errorf(midas.ErrInvalid, "event %s is invalid", h.Payload.Event()))
		return
	}
}

func (h StrapiToHugoHandler) handleCreateSingle(w http.ResponseWriter, r *http.Request) {
	if _, err := h.HugoSite.UpdateSingle(h.Payload); err != nil {
		Error(w, r, err)
		return
	}

	if err := h.HugoSite.BuildSite(true); err != nil {
		Error(w, r, err)
		return
	}

	if err := h.runDeploys(r); err != nil {
		Error(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h StrapiToHugoHandler) handleCreateCollection(w http.ResponseWriter, r *http.Request) {
	if _, err := h.HugoSite.CreateEntry(h.Payload); err != nil {
		Error(w, r, err)
		return
	}

	if err := h.HugoSite.BuildSite(true); err != nil {
		Error(w, r, err)
		return
	}

	if err := h.runDeploys(r); err != nil {
		Error(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h StrapiToHugoHandler) handleUpdateSingle(w http.ResponseWriter, r *http.Request) {
	if _, err := h.HugoSite.UpdateSingle(h.Payload); err != nil {
		Error(w, r, err)
		return
	}

	if err := h.HugoSite.BuildSite(true); err != nil {
		Error(w, r, err)
		return
	}

	if err := h.runDeploys(r); err != nil {
		Error(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h StrapiToHugoHandler) handleUpdateCollection(w http.ResponseWriter, r *http.Request) {
	if _, err := h.HugoSite.UpdateEntry(h.Payload); err != nil {
		Error(w, r, err)
		return
	}

	if err := h.HugoSite.BuildSite(true); err != nil {
		Error(w, r, err)
		return
	}

	if err := h.runDeploys(r); err != nil {
		Error(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h StrapiToHugoHandler) handleDeleteCollection(w http.ResponseWriter, r *http.Request) {
	if _, err := h.HugoSite.DeleteEntry(h.Payload); err != nil {
		Error(w, r, err)
		return
	}

	if err := h.HugoSite.BuildSite(true); err != nil {
		Error(w, r, err)
		return
	}

	if err := h.runDeploys(r); err != nil {
		Error(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// runDeploys executes both final and the draft deploys.
func (h StrapiToHugoHandler) runDeploys(r *http.Request) error {
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
func (h StrapiToHugoHandler) deploy(cfg *midas.Site, draft bool) error {
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

		if deploymentService, err = dpl(*cfg, dplSettings); err != nil {
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
