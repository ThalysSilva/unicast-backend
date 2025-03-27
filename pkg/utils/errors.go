package utils

import (
	"errors"
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
)

func TraceError(context string, err error) error {
	return fmt.Errorf("%s: %w", context, err)
}

type CustomError struct {
	HttpCode int
	Err      error
}

func (e *CustomError) Error() string {
	return e.Err.Error()
}

func (e *CustomError) MakeError(message string, httpCode int) *CustomError {
	return &CustomError{
		HttpCode: httpCode,
		Err:      errors.New(message),
	}
}

func HandleErrorResponse(g *gin.Context, err error) {
	defer g.Abort()
	customErr := &CustomError{}
	if errors.As(err, &customErr) {
		g.JSON(customErr.HttpCode, gin.H{"error": customErr.Error()})
		return
	}
	errString := fmt.Sprintf("Erro interno inesperado: %s", err.Error())
	fmt.Println(errString) // debug
	g.JSON(http.StatusInternalServerError, gin.H{"error": errString})

}
