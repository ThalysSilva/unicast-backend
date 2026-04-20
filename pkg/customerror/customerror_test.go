package customerror

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandleResponseUsesPublicMessageForCustomError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		HandleResponse(c, Make("mensagem pública", http.StatusBadRequest, errors.New("internal detail")))
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/test", nil))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if recorder.Body.String() != `{"error":"mensagem pública"}` {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestHandleResponseHidesUnexpectedErrorDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		HandleResponse(c, errors.New("database password leaked in error"))
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/test", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if recorder.Body.String() != `{"error":"Erro interno inesperado"}` {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}
