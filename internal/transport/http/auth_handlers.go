package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_JSON", "request body is invalid", nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "request fields are invalid", validationFields(err))
		return
	}
	tokens, err := h.auth.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		writeError(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "email or password is invalid", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tokens})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (h *Handler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_JSON", "request body is invalid", nil)
		return
	}
	tokens, err := h.auth.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		writeError(c, http.StatusUnauthorized, "REFRESH_REJECTED", "refresh token is expired, revoked or already used", nil)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tokens})
}
