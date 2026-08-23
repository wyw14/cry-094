package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/wyw14/cry-094/internal/domain/common"
)

func writeError(c *gin.Context, status int, code, message string, fields []common.FieldError) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message, "fields": fields, "request_id": c.GetString("request_id")}})
}
func handleError(c *gin.Context, err error) {
	var coded *common.CodedError
	if errors.As(err, &coded) {
		writeError(c, http.StatusBadRequest, coded.Code, coded.Message, coded.Fields)
		return
	}
	if errors.Is(err, common.ErrForbidden) {
		writeError(c, http.StatusForbidden, "FORBIDDEN", err.Error(), nil)
		return
	}
	if errors.Is(err, common.ErrNotFound) {
		writeError(c, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
		return
	}
	if errors.Is(err, common.ErrConflict) || errors.Is(err, common.ErrVersionConflict) {
		writeError(c, http.StatusConflict, "CONFLICT", err.Error(), nil)
		return
	}
	if errors.Is(err, common.ErrInvalidTransition) {
		writeError(c, http.StatusConflict, "INVALID_TRANSITION", err.Error(), nil)
		return
	}
	writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "request could not be completed", nil)
}
func validationFields(err error) []common.FieldError {
	var values validator.ValidationErrors
	if !errors.As(err, &values) {
		return nil
	}
	fields := make([]common.FieldError, 0, len(values))
	for _, value := range values {
		fields = append(fields, common.FieldError{Field: value.Field(), Message: "value failed " + value.Tag() + " validation"})
	}
	return fields
}
