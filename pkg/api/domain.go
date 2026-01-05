package api

type ErrorResponse struct {
	Error string `json:"error"`
}

type DefaultResponse[T any] struct {
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// MessageResponse cobre respostas sem payload de dados.
type MessageResponse struct {
	Message string `json:"message"`
}
