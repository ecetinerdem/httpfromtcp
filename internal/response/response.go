package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type WriterStatus string

const (
	writeStatusLine WriterStatus = "writeStatusLine"
	writeHeaders    WriterStatus = "writeHeaders"
	writeBody       WriterStatus = "writeBody"
)

type Writer struct {
	W            io.Writer
	WriterStatus WriterStatus
}

type StatusCode int

const (
	StatusCodeOK             StatusCode = 200
	StatusCodeBadRequest     StatusCode = 400
	StatusCodeInternalServer StatusCode = 500
)

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	switch statusCode {
	case StatusCodeOK:
		_, err := w.W.Write([]byte(fmt.Sprintf("HTTP/1.1 %d OK\r\n", StatusCodeOK)))
		if err != nil {
			return err
		}
	case StatusCodeBadRequest:
		_, err := w.W.Write([]byte(fmt.Sprintf("HTTP/1.1 %d Bad Request\r\n", StatusCodeBadRequest)))
		if err != nil {
			return err
		}
	case StatusCodeInternalServer:
		_, err := w.W.Write([]byte(fmt.Sprintf("HTTP/1.1 %d Internal Server Error\r\n", StatusCodeInternalServer)))
		if err != nil {
			return err
		}
	default:
		_, err := w.W.Write([]byte(fmt.Sprintf("HTTP/1.1 %d ", statusCode)))
		if err != nil {
			return err
		}
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	for key, value := range h {
		_, err := fmt.Fprintf(w.W, "%s: %s\r\n", key, value)
		if err != nil {
			return err
		}
	}
	w.W.Write([]byte("\r\n"))
	return nil
}

func (w *Writer) WriteBody(b []byte) (int, error) {
	n, err := w.W.Write(b)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	err := w.WriteHeaders(h)

	if err != nil {
		return err
	}
	w.W.Write([]byte("\r\n"))
	return nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	total := 0

	// Write chunk size in hexadecimal
	n, err := w.WriteBody([]byte(fmt.Sprintf("%x\r\n", len(p))))
	total += n
	if err != nil {
		return total, err
	}

	// Write the actual chunk data
	n, err = w.WriteBody(p)
	total += n
	if err != nil {
		return total, err
	}

	// End the chunk
	n, err = w.WriteBody([]byte("\r\n"))
	total += n
	if err != nil {
		return total, err
	}

	return total, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	return w.WriteBody([]byte("0\r\n"))
}
