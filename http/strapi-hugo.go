package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/kovansky/midas"
	"github.com/kovansky/midas/hugo"
	"github.com/kovansky/midas/strapi"
	"io"
	"net/http"
)

type StrapiToHugoHandler struct {
	HugoSite midas.SiteService
	Payload  midas.Payload
}

func (s *Server) registerStrapiToHugoRoutes(r chi.Router) {
	r.Post("/strapi/hugo", s.handleStrapiToHugo)
}

func (s *Server) handleStrapiToHugo(w http.ResponseWriter, r *http.Request) {
	cfg := midas.UserConfigFromContext(r.Context()).(midas.Site)

	fmt.Println("Received")

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

	handler := &StrapiToHugoHandler{
		HugoSite: hugo.NewSiteService(cfg),
		Payload:  payload,
	}

	handler.Handle(w, r)
}

func (h StrapiToHugoHandler) Handle(w http.ResponseWriter, r *http.Request) {
	cfg := midas.UserConfigFromContext(r.Context()).(midas.Site)

	model := h.Payload.Metadata()["model"].(string)
	var isSingle bool

	if midas.Contains(cfg.SingleTypes, model) {
		isSingle = true
	} else if midas.Contains(cfg.CollectionTypes, model) {
		isSingle = false
	} else {
		Error(w, r, midas.Errorf(midas.ErrUnaccepted, "model %s is not accepted", model))
		return
	}

	switch h.Payload.Event() {
	case strapi.Create.String():
		if isSingle {
			h.handleCreateSingle(w, r)
		} else {
			h.handleCreateCollection(w, r)
		}
		break
	case strapi.Update.String():
		if isSingle {
			h.handleUpdateSingle(w, r)
		}
		break
	default:
		Error(w, r, midas.Errorf(midas.ErrInvalid, "event %s is invalid", h.Payload.Event()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h StrapiToHugoHandler) handleCreateSingle(w http.ResponseWriter, r *http.Request) {
	err := h.HugoSite.BuildSite(true)

	if err != nil {
		Error(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h StrapiToHugoHandler) handleCreateCollection(w http.ResponseWriter, r *http.Request) {
	_, err := h.HugoSite.CreateEntry(h.Payload)

	if err != nil {
		Error(w, r, err)
		return
	}

	err = h.HugoSite.BuildSite(true)

	if err != nil {
		Error(w, r, err)
		return
	}

	// ToDo: add to registry

	w.WriteHeader(http.StatusNoContent)
}

func (h StrapiToHugoHandler) handleUpdateSingle(w http.ResponseWriter, r *http.Request) {
	err := h.HugoSite.BuildSite(false)

	if err != nil {
		Error(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
