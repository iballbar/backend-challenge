package http

import (
	"errors"
	"fmt"
	"net/http"

	"backend-challenge/internal/core/domain"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

func writeJSON(c *gin.Context, status int, payload any) {
	c.JSON(status, payload)
}

func writeError(c *gin.Context, status int, message string) {
	writeJSON(c, status, ErrorResponse{Error: message})
}

func writeValidationError(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		details := make(map[string]string)
		for _, fe := range ve {
			details[fe.Field()] = msgForTag(fe.Tag(), fe.Param())
		}
		writeJSON(c, http.StatusBadRequest, ErrorResponse{
			Error:   "validation failed",
			Details: details,
		})
		return
	}
	writeError(c, http.StatusBadRequest, "invalid request payload")
}

func msgForTag(tag, param string) string {
	switch tag {
	case "required":
		return "this field is required"
	case "email":
		return "invalid email format"
	case "min":
		return fmt.Sprintf("minimum length is %s", param)
	case "len":
		return fmt.Sprintf("length must be exactly %s", param)
	case "max":
		return fmt.Sprintf("maximum value/length is %s", param)
	}
	return "invalid value"
}

func handleError(c *gin.Context, err error) {
	switch err {
	case domain.ErrInvalidCredentials:
		writeError(c, http.StatusUnauthorized, err.Error())
	case domain.ErrEmailAlreadyExists, domain.ErrInvalidEmail:
		writeError(c, http.StatusBadRequest, err.Error())
	case domain.ErrNotFound:
		writeError(c, http.StatusNotFound, err.Error())
	case domain.ErrInvalidPattern:
		writeError(c, http.StatusBadRequest, err.Error())
	case domain.ErrNoTicketsAvailable:
		writeError(c, http.StatusNotFound, err.Error())
	default:
		writeError(c, http.StatusInternalServerError, "internal server error")
	}
}
