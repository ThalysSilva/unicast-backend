package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ThalysSilva/unicast-backend/pkg/api"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationErrorHandler é um middleware que intercepta erros de validação e os transforma em respostas JSON
func ValidationErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		errMessages := make(map[string]string)

		for _, e := range c.Errors {
			accumulateValidationErrors(errMessages, e.Err)
		}

		if len(errMessages) > 0 {
			payload, _ := json.Marshal(errMessages)
			c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: string(payload)})
			c.Abort()
			return
		}
	}
}

func accumulateValidationErrors(errMessages map[string]string, err error) {
	switch typed := err.(type) {
	case validator.ValidationErrors:
		for _, fe := range typed {
			errMessages[fe.Field()] = validationMessage(fe)
		}
	case *json.SyntaxError:
		errMessages["general"] = "JSON inválido: sintaxe incorreta"
	}
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("O campo %s é obrigatório", fe.Field())
	case "email":
		return fmt.Sprintf("O campo %s deve ser um email válido", fe.Field())
	default:
		return fmt.Sprintf("Erro de validação no campo %s", fe.Field())
	}
}
