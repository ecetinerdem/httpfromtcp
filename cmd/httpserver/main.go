package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"httpfromtcp/cmd/httpserver/server"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

const port = 42069

func main() {

	server, err := server.Serve(port, handlerFunc)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handlerFunc(writer *response.Writer, req *request.Request) {

	switch {
	case req.RequestLine.RequestTarget == "/":
		html := `
			<html>
  			<head>
    			<title>200 OK</title>
  			</head>
  			<body>
    			<h1>Success!</h1>
    			<p>Your request was an absolute banger.</p>
  			</body>
			</html>
		`
		headersToSend := response.GetDefaultHeaders(len(html))
		headersToSend.Replace("Content-Type", "text/html")
		writer.WriteStatusLine(response.StatusCodeOK)
		writer.WriteHeaders(headersToSend)
		writer.WriteBody([]byte(html))
		return

	case req.RequestLine.RequestTarget == "/yourproblem":
		html := `
			<html>
  			<head>
    			<title>400 Bad Request</title>
  			</head>
  			<body>
    			<h1>Bad Request</h1>
    			<p>Your request honestly kinda sucked.</p>
  			</body>
			</html>
		`
		headersToSend := response.GetDefaultHeaders(len(html))
		headersToSend.Replace("Content-Type", "text/html")
		writer.WriteStatusLine(response.StatusCodeBadRequest)
		writer.WriteHeaders(headersToSend)
		writer.WriteBody([]byte(html))
		return
	case req.RequestLine.RequestTarget == "/myproblem":
		html := `
			<html>
  			<head>
    			<title>500 Internal Server Error</title>
  			</head>
  			<body>
    			<h1>Internal Server Error</h1>
    			<p>Okay, you know what? This one is on me.</p>
  			</body>
			</html>
		`
		headersToSend := response.GetDefaultHeaders(len(html))
		headersToSend.Replace("Content-Type", "text/html")
		writer.WriteStatusLine(response.StatusCodeInternalServer)
		writer.WriteHeaders(headersToSend)
		writer.WriteBody([]byte(html))
		return

	case strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/"):
		target := req.RequestLine.RequestTarget

		path := strings.TrimPrefix(target, "/httpbin/")

		resp, err := http.Get("https://httpbingo.org/" + path)
		if err != nil {
			html := `
	<html>
		<head>
			<title>500 Internal Server Error</title>
		</head>
		<body>
			<h1>Internal Server Error</h1>
			<p>Okay, you know what? This one is on me.</p>
		</body>
	</html>
	`

			headersToSend := response.GetDefaultHeaders(len(html))
			headersToSend.Replace("Content-Type", "text/html")

			writer.WriteStatusLine(response.StatusCodeInternalServer)
			writer.WriteHeaders(headersToSend)
			writer.WriteBody([]byte(html))
			return
		}

		defer resp.Body.Close()

		writer.WriteStatusLine(response.StatusCodeOK)

		headersToSend := response.GetDefaultHeaders(0)
		headersToSend.Delete("Content-Length")
		headersToSend.Set("Transfer-Encoding", "chunked")
		headersToSend.Replace("Content-Type", "text/plain")
		headersToSend.Set("Trailer", "X-Content-SHA256, X-Content-Length")

		writer.WriteHeaders(headersToSend)

		respBody := make([]byte, 0)

		for {
			data := make([]byte, 32)

			n, err := resp.Body.Read(data)

			// If we successfully received data, write it as a chunk
			if n > 0 {
				_, err := writer.WriteChunkedBody(data[:n])
				if err != nil {
					return
				}
			}

			respBody = append(respBody, data[:n]...)
			// EOF means we've finished reading the response
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				return
			}
		}

		// Tell the client that the chunked body is finished
		_, err = writer.WriteChunkedBodyDone()
		if err != nil {
			return
		}

		var trailers headers.Headers = make(map[string]string)

		trailerLen := len(respBody)
		trailerHash := sha256.Sum256(respBody)
		trailers.Set("X-Content-SHA256", hex.EncodeToString(trailerHash[:]))
		trailers.Set("X-Content-Length", strconv.Itoa(trailerLen))
		err = writer.WriteTrailers(trailers)
		if err != nil {
			log.Println("Trailer err: ", err)
		}

		return
	case req.RequestLine.RequestTarget == "/video":
		var headersToSend headers.Headers = make(map[string]string)
		headersToSend.Set("Content-Type", "video/mp4")
		mp4, err := os.ReadFile("assets/vim.mp4")
		if err != nil {
			log.Println("File err: ", err)
			return
		}
		writer.WriteStatusLine(response.StatusCodeOK)
		writer.WriteHeaders(headersToSend)
		writer.WriteBody(mp4)
		return
	default:
		html := `
		<html>
			<head>
				<title>500 Internal Server Error</title>
			</head>
			<body>
				<h1>Internal Server Error</h1>
				<p>Okay, you know what? This one is on me.</p>
			</body>
		</html>
		`
		headersToSend := response.GetDefaultHeaders(len(html))
		headersToSend.Replace("Content-Type", "text/html")
		writer.WriteStatusLine(response.StatusCodeInternalServer)
		writer.WriteHeaders(headersToSend)
		writer.WriteBody([]byte(html))
	}

}
