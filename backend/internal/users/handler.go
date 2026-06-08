package users

import (
	"net/http"

	"github.com/luizf/go-chat/backend/internal/apperr"
	"github.com/luizf/go-chat/backend/internal/httpx"
	"github.com/luizf/go-chat/backend/internal/requestctx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := requestctx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, apperr.ErrUnauthorized)
		return
	}
	u, err := h.svc.GetProfile(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, ToResponse(u))
}

func (h *Handler) SearchByUsername(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		httpx.WriteError(w, apperr.ErrInvalidInput)
		return
	}
	u, err := h.svc.SearchByUsername(r.Context(), username)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, ToSearchResponse(u))
}

type Response struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Username  string  `json:"username"`
	AvatarURL *string `json:"avatar_url"`
}

type SearchResponse struct {
	ID        string  `json:"id"`
	Username  string  `json:"username"`
	AvatarURL *string `json:"avatar_url"`
}

func ToResponse(u *User) Response {
	return Response{
		ID:        u.ID.String(),
		Email:     u.Email,
		Username:  u.Username,
		AvatarURL: u.AvatarURL,
	}
}

func ToSearchResponse(u *User) SearchResponse {
	return SearchResponse{
		ID:        u.ID.String(),
		Username:  u.Username,
		AvatarURL: u.AvatarURL,
	}
}
