package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sneakers-store/internal/clients/sso/grpc"
	"sneakers-store/internal/logger/sl"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	ssoClient *grpc.Client
	log       *slog.Logger
}

func New(ssoClient *grpc.Client, log *slog.Logger) *Handler {
	return &Handler{ssoClient: ssoClient, log: log}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	AppId    int32  `json:"app_id"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	token, err := h.ssoClient.Login(r.Context(), req.Email, req.Password, req.AppId)
	if err != nil {
		h.log.Error("Login failed", sl.Err(err))
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	response := map[string]string{"token": token}
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	UserID, err := h.ssoClient.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		h.log.Error("Registration failed", sl.Err(err))

		if status.Code(err) == codes.AlreadyExists {
			http.Error(w, "User already exist", http.StatusConflict)
		}

		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": UserID,
		"message": "User registered successfully",
	})
}
