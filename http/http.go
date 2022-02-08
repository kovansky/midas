package http

import (
	"encoding/json"
	"github.com/go-chi/httplog"
	"github.com/kovansky/midas"
	"net/http"
)

// Error prints, and optionally logs, an error message.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	// Extract error code and message
	code, message := midas.ErrorCode(err), midas.ErrorMessage(err)

	// Log internal errors
	if code == midas.ErrInternal {
		midas.ReportError(r.Context(), err, r)

		message = "Internal server error" // We don't want the error to be displayed for the enduser
	}

	jsonError, _ := json.Marshal(&ErrorResponse{Error: message})

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(ErrorStatusCode(code))
	_, _ = w.Write(jsonError)

	log := httplog.LogEntry(r.Context())
	log.Error().Err(err).Msg("details of following errored request")
}

type ErrorResponse struct {
	Error string `json:"error"`
}

var codes = map[string]int{
	midas.ErrUnauthorized: http.StatusUnauthorized,
	midas.ErrInvalid:      http.StatusBadRequest,
	midas.ErrUnaccepted:   http.StatusBadRequest,
	midas.ErrInternal:     http.StatusInternalServerError,
	midas.ErrRegistry:     http.StatusInternalServerError,
	midas.ErrSiteConfig:   http.StatusInternalServerError,
}

func ErrorStatusCode(code string) int {
	if httpCode, ok := codes[code]; ok {
		return httpCode
	}

	return http.StatusInternalServerError
}
