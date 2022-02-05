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
		log := httplog.LogEntry(r.Context())
		log.Error().Msgf("internal server error: %+v", err)
		midas.ReportError(r.Context(), err, r)
	}

	jsonError, _ := json.Marshal(&ErrorResponse{Error: message})

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(ErrorStatusCode(code))
	_, _ = w.Write(jsonError)
}

type ErrorResponse struct {
	Error string `json:"error"`
}

var codes = map[string]int{
	midas.ErrUnauthorized: http.StatusUnauthorized,
	midas.ErrInvalid:      http.StatusBadRequest,
	midas.ErrUnaccepted:   http.StatusBadRequest,
	midas.ErrInternal:     http.StatusInternalServerError,
}

func ErrorStatusCode(code string) int {
	if httpCode, ok := codes[code]; ok {
		return httpCode
	}

	return http.StatusInternalServerError
}
