package response

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Response struct {
	IsSuccess bool   `json:"is_success"`
	Message   string `json:"message"`
	Data      any    `json:"data"`
}

type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func getCustomMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", fe.Field(), fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s characters long", fe.Field(), fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s characters long", fe.Field(), fe.Param())
	case "eqfield":
		return fmt.Sprintf("%s must match %s", fe.Field(), fe.Param())
	}
	return fe.Error() // Default error
}

// HandleBindError processes bind/validation errors and writes the standard JSON response.
// Returns true if an error was found and handled.
func HandleBindError(ctx *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		out := make([]ErrorDetail, len(ve))
		for i, fe := range ve {
			out[i] = ErrorDetail{
				Field:   fe.Field(),
				Message: getCustomMessage(fe),
			}
		}

		ctx.JSON(http.StatusUnprocessableEntity, Response{
			IsSuccess: false,
			Message:   "Validation failed",
			Data:      out,
		})
		return true
	}

	// Fallback for other binding errors (e.g., malformed JSON)
	ctx.JSON(http.StatusBadRequest, Response{
		IsSuccess: false,
		Message:   "Invalid request format",
		Data:      err.Error(),
	})
	return true
}

// Success responds with a standard success payload
func Success(ctx *gin.Context, statusCode int, message string, data any) {
	ctx.JSON(statusCode, Response{
		IsSuccess: true,
		Message:   message,
		Data:      data,
	})
}
