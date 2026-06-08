package httpx

import (
	"context"
	"encoding/json"
	"net/http"
)

// Validator is implemented by request DTOs with field-level validation.
type Validator interface {
	Valid(ctx context.Context) map[string]string
}

func DecodeValid[T Validator](r *http.Request) (T, map[string]string, error) {
	var v T
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&v); err != nil {
		return v, nil, err
	}
	if problems := v.Valid(r.Context()); len(problems) > 0 {
		return v, problems, nil
	}
	return v, nil, nil
}

func WriteValidationError(w http.ResponseWriter, problems map[string]string) {
	WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
		"error":  "invalid input",
		"fields": problems,
	})
}
