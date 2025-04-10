package customerror

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Trace(context string, err error) error {
	return fmt.Errorf("%s: %w", context, err)
}

type CustomError struct {
	HttpCode int
	message  string
	Err      error
}

func (e *CustomError) Error() string {
	fmt.Println("Preparando para mostrar o erro")
	fmt.Println("CustomError", e.message)
	return fmt.Sprintf("%s: %s", e.message, e.Err.Error())
}

func Make(message string, httpCode int, err error) *CustomError {
	return &CustomError{
		HttpCode: httpCode,
		Err:      err,
		message:  message,
	}
}

func HandleResponse(g *gin.Context, err error) {
	defer g.Abort()
	customErr := &CustomError{}
	if errors.As(err, &customErr) {
		g.JSON(customErr.HttpCode, gin.H{"error": customErr.Error()})
		return
	}
	errString := fmt.Sprintf("Erro interno inesperado: %s", err.Error())
	g.JSON(http.StatusInternalServerError, gin.H{"error": errString})

}
