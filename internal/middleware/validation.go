package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func ValidationErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // Executa o handler

		if len(c.Errors) == 0 {
			return
		}

		errMessages := make(map[string]string)

		for _, e := range c.Errors {
			switch err := e.Err.(type) {
			case validator.ValidationErrors:
				// Erros de validação (ex.: campo required, email inválido)
				for _, fe := range err {
					switch fe.Tag() {
					case "required":
						errMessages[fe.Field()] = fmt.Sprintf("O campo %s é obrigatório", fe.Field())
					case "email":
						errMessages[fe.Field()] = fmt.Sprintf("O campo %s deve ser um email válido", fe.Field())
					default:
						errMessages[fe.Field()] = fmt.Sprintf("Erro de validação no campo %s", fe.Field())
					}
				}
			case *json.SyntaxError:
				// Erros de parsing de JSON (ex.: vírgula extra)
				errMessages["general"] = "JSON inválido: sintaxe incorreta"
			default:
				// Outros erros são ignorados por este middleware
				continue
			}
		}

		if len(errMessages) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"errors": errMessages})
			c.Abort()
			return
		}
	}
}