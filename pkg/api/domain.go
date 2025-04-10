package api

type ErrorResponse struct {
	Error string `json:"error"`
}

type DefaultResponse[T any] struct {
	Message string `json:"message"`
	Data    T      `json:"data"`
}
