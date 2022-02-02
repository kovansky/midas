package http

import (
	"encoding/json"
	"fmt"
	"github.com/kovansky/midas"
	"net/http"
)

// Error prints, and optionally logs, an error message.
func Error(w http.ResponseWriter, _ *http.Request, err error) {
	// Extract error code and message
	code, message := midas.ErrorCode(err), midas.ErrorMessage(err)

	// Log internal errors
	if code == midas.ErrInternal {
		fmt.Printf("%+v", err)
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
