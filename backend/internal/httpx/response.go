package httpx

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/luizf/go-chat/backend/internal/apperr"
)

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, apperr.ErrInvalidInput):
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, apperr.ErrUnauthorized), errors.Is(err, apperr.ErrInvalidCredentials):
		WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	case errors.Is(err, apperr.ErrForbidden):
		WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
	case errors.Is(err, apperr.ErrNotFound):
		WriteJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	case errors.Is(err, apperr.ErrConflict):
		WriteJSON(w, http.StatusConflict, map[string]string{"error": "conflict"})
	default:
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

func DecodeJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
