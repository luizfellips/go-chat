package httpx_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/luizf/go-chat/backend/internal/apperr"
	"github.com/luizf/go-chat/backend/internal/httpx"
)

func TestWriteError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "invalid input",
			err:        apperr.ErrInvalidInput,
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"invalid input"}` + "\n",
		},
		{
			name:       "unauthorized",
			err:        apperr.ErrUnauthorized,
			wantStatus: http.StatusUnauthorized,
			wantBody:   `{"error":"unauthorized"}` + "\n",
		},
		{
			name:       "invalid credentials",
			err:        apperr.ErrInvalidCredentials,
			wantStatus: http.StatusUnauthorized,
			wantBody:   `{"error":"unauthorized"}` + "\n",
		},
		{
			name:       "forbidden",
			err:        apperr.ErrForbidden,
			wantStatus: http.StatusForbidden,
			wantBody:   `{"error":"forbidden"}` + "\n",
		},
		{
			name:       "not found",
			err:        apperr.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   `{"error":"not found"}` + "\n",
		},
		{
			name:       "conflict",
			err:        apperr.ErrConflict,
			wantStatus: http.StatusConflict,
			wantBody:   `{"error":"conflict"}` + "\n",
		},
		{
			name:       "unknown error",
			err:        errors.New("database exploded"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"internal server error"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			httpx.WriteError(rec, tt.err)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := rec.Body.String(); got != tt.wantBody {
				t.Fatalf("body = %q, want %q", got, tt.wantBody)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Fatalf("Content-Type = %q, want application/json", ct)
			}
		})
	}
}
