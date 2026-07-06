package utils

type JSONResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

func NewSuccess(message string, data any) JSONResponse {
	return JSONResponse{Message: message, Data: data}
}

func NewList(message string, data any, meta any) JSONResponse {
	return JSONResponse{Message: message, Data: data, Meta: meta}
}

func NewMessage(message string) JSONResponse {
	return JSONResponse{Message: message}
}
