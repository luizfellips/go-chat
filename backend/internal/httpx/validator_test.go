package httpx_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/luizf/go-chat/backend/internal/httpx"
)

type sampleRequest struct {
	Name string `json:"name"`
}

func (r sampleRequest) Valid(_ context.Context) map[string]string {
	if strings.TrimSpace(r.Name) == "" {
		return map[string]string{"name": "required"}
	}
	return nil
}

func TestDecodeValidSuccess(t *testing.T) {
	t.Parallel()
	body := bytes.NewBufferString(`{"name":"alice"}`)
	r := httptest.NewRequest(http.MethodPost, "/", body)

	got, problems, err := httpx.DecodeValid[sampleRequest](r)
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) > 0 {
		t.Fatalf("problems = %v", problems)
	}
	if got.Name != "alice" {
		t.Fatalf("name = %q", got.Name)
	}
}

func TestDecodeValidFieldProblems(t *testing.T) {
	t.Parallel()
	body := bytes.NewBufferString(`{"name":""}`)
	r := httptest.NewRequest(http.MethodPost, "/", body)

	_, problems, err := httpx.DecodeValid[sampleRequest](r)
	if err != nil {
		t.Fatal(err)
	}
	if problems["name"] != "required" {
		t.Fatalf("problems = %v", problems)
	}
}

func TestWriteValidationError(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	httpx.WriteValidationError(rec, map[string]string{"email": "required"})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"fields"`) {
		t.Fatalf("body = %q", rec.Body.String())
	}
}
