package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	listener net.Listener
	handler  Handler
	closed   atomic.Bool
}

func Serve(n int, handlerFunc Handler) (*Server, error) {

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", n))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
		handler:  handlerFunc,
	}
	server.listen()

	return server, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {

	for {

		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			return
		}
		go s.handle(conn)

	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	writer := &response.Writer{}
	writer.W = conn
	if err != nil {
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
		headers := response.GetDefaultHeaders(len(html))
		headers.Replace("Content-Type", "text/html")
		writer.WriteStatusLine(response.StatusCodeBadRequest)
		writer.WriteHeaders(headers)
		writer.WriteBody([]byte(html))
		return
	}

	s.handler(writer, req)
}
