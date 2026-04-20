package customerror

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/ThalysSilva/unicast-backend/pkg/api"
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
	return fmt.Sprintf("%s: %s", e.message, e.Err.Error())
}

func (e *CustomError) PublicMessage() string {
	return e.message
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
	log.Printf("request error method=%s path=%s error=%v", g.Request.Method, g.Request.URL.Path, err)

	customErr := &CustomError{}
	if errors.As(err, &customErr) {
		g.JSON(customErr.HttpCode, api.ErrorResponse{Error: customErr.PublicMessage()})
		return
	}
	g.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: "Erro interno inesperado"})

}
