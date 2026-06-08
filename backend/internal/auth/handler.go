package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/luizf/go-chat/backend/internal/apperr"
	"github.com/luizf/go-chat/backend/internal/httpx"
	"github.com/luizf/go-chat/backend/internal/users"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type registerRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r registerRequest) Valid(_ context.Context) map[string]string {
	problems := make(map[string]string)
	email := strings.TrimSpace(strings.ToLower(r.Email))
	if email == "" {
		problems["email"] = "required"
	} else if !strings.Contains(email, "@") || len(email) < 3 {
		problems["email"] = "invalid format"
	}
	username := strings.TrimSpace(r.Username)
	if username == "" {
		problems["username"] = "required"
	} else if len(username) > 50 {
		problems["username"] = "must be at most 50 characters"
	}
	if len(r.Password) < 8 {
		problems["password"] = "must be at least 8 characters"
	}
	return problems
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	req, problems, err := httpx.DecodeValid[registerRequest](r)
	if err != nil {
		httpx.WriteError(w, apperr.ErrInvalidInput)
		return
	}
	if len(problems) > 0 {
		httpx.WriteValidationError(w, problems)
		return
	}
	user, err := h.svc.Register(r.Context(), RegisterInput{
		Email: req.Email, Username: req.Username, Password: req.Password,
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, map[string]interface{}{"user": users.ToResponse(user)})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r loginRequest) Valid(_ context.Context) map[string]string {
	problems := make(map[string]string)
	if strings.TrimSpace(r.Email) == "" {
		problems["email"] = "required"
	}
	if r.Password == "" {
		problems["password"] = "required"
	}
	return problems
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	req, problems, err := httpx.DecodeValid[loginRequest](r)
	if err != nil {
		httpx.WriteError(w, apperr.ErrInvalidInput)
		return
	}
	if len(problems) > 0 {
		httpx.WriteValidationError(w, problems)
		return
	}
	out, err := h.svc.Login(r.Context(), LoginInput{Email: req.Email, Password: req.Password})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  out.Tokens.AccessToken,
		"refresh_token": out.Tokens.RefreshToken,
		"expires_in":    out.Tokens.ExpiresIn,
		"user":          users.ToResponse(out.User),
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (r refreshRequest) Valid(_ context.Context) map[string]string {
	problems := make(map[string]string)
	if strings.TrimSpace(r.RefreshToken) == "" {
		problems["refresh_token"] = "required"
	}
	return problems
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	req, problems, err := httpx.DecodeValid[refreshRequest](r)
	if err != nil {
		httpx.WriteError(w, apperr.ErrInvalidInput)
		return
	}
	if len(problems) > 0 {
		httpx.WriteValidationError(w, problems)
		return
	}
	out, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  out.Tokens.AccessToken,
		"refresh_token": out.Tokens.RefreshToken,
		"expires_in":    out.Tokens.ExpiresIn,
		"user":          users.ToResponse(out.User),
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	req, problems, err := httpx.DecodeValid[refreshRequest](r)
	if err != nil {
		httpx.WriteError(w, apperr.ErrInvalidInput)
		return
	}
	if len(problems) > 0 {
		httpx.WriteValidationError(w, problems)
		return
	}
	if err := h.svc.Logout(r.Context(), req.RefreshToken); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
