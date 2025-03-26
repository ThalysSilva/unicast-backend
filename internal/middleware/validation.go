package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)
// ValidationErrorHandler é um middleware que intercepta erros de validação e os transforma em respostas JSON
func ValidationErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() 

		if len(c.Errors) == 0 {
			return
		}

		errMessages := make(map[string]string)

		for _, e := range c.Errors {
			switch err := e.Err.(type) {
			case validator.ValidationErrors:
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
				errMessages["general"] = "JSON inválido: sintaxe incorreta"
			default:
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