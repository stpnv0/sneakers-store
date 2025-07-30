package auth

import (
	"api_gateway/internal/client/sso"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const AppID = 1

type Handler struct {
	client *sso.Client
	log    *slog.Logger
}

func NewHandler(client *sso.Client, log *slog.Logger) *Handler {
	return &Handler{client: client, log: log}
}

// Register обрабатывает POST /auth/register
func (h *Handler) Register(c *gin.Context) {
	var reqBody struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=3"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := h.client.Register(c.Request.Context(), reqBody.Email, reqBody.Password)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": "user with this email already exists"})
			return
		}
		h.log.Error("failed to register user", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user_id": userID})
}

// Login обрабатывает POST /auth/login
func (h *Handler) Login(c *gin.Context) {
	var reqBody struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.client.Login(c.Request.Context(), reqBody.Email, reqBody.Password, AppID)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.InvalidArgument {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email or password"})
			return
		}
		h.log.Error("failed to login user", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
