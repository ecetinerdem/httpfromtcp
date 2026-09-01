package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w *response.Writer, req *request.Request)

func WriteHandlerError(e HandlerError, w io.Writer) error {
	_, err := fmt.Fprintf(w, "%v", e)
	if err != nil {
		return err
	}
	return nil
}
