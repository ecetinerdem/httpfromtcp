# HTTP From TCP

A lightweight HTTP/1.1 server built from scratch in Go using raw TCP connections.

This project implements the fundamental components of the HTTP protocol without relying on Go's built-in HTTP server for incoming requests. It focuses on understanding how HTTP/1.1 works at the protocol level by manually parsing requests, managing headers, constructing responses, handling request bodies, and implementing chunked transfer encoding and trailers.

The project was built as a learning exercise to explore the relationship between TCP streams and the HTTP application protocol.

## Features

* HTTP/1.1 request parsing
* Raw TCP server using `net.Listener`
* Manual request line parsing
* HTTP header parsing and validation
* Case-insensitive header handling
* Request body parsing using `Content-Length`
* HTTP response generation
* Status line generation
* Response header writing
* Response body writing
* Chunked transfer encoding
* HTTP trailers
* SHA-256 response integrity trailer
* Proxying requests to `httpbingo.org`
* Static MP4 video serving
* Graceful server shutdown using OS signals

## Architecture

The project is organized into separate packages responsible for different parts of the HTTP lifecycle.

```text
.
├── cmd/
│   └── httpserver/
│       ├── main.go
│       └── server/
│
├── internal/
│   ├── headers/
│   │   └── headers.go
│   │
│   ├── request/
│   │   └── request.go
│   │
│   └── response/
│       └── response.go
│
└── assets/
    └── vim.mp4
```

The general request lifecycle is:

```text
Client
  │
  │ TCP Connection
  ▼
net.Listener
  │
  ▼
Server
  │
  ▼
Request Parser
  │
  ├── Request Line
  ├── Headers
  └── Body
  │
  ▼
Handler
  │
  ▼
Response Writer
  │
  ├── Status Line
  ├── Headers
  ├── Body
  ├── Chunked Body
  └── Trailers
  │
  ▼
Client
```

# HTTP Request Parsing

Incoming requests are parsed directly from an `io.Reader`.

The request parser does not assume that an entire HTTP request arrives in a single TCP read. Since TCP is a stream protocol, request data can arrive in multiple pieces.

The parser maintains an internal state machine:

```text
initialized
     │
     ▼
parsing headers
     │
     ▼
parsing body
     │
     ▼
done
```

The supported request structure is:

```text
GET / HTTP/1.1
Host: localhost

```

The request line is parsed into:

```go
type RequestLine struct {
    HttpVersion   string
    RequestTarget string
    Method        string
}
```

The parser validates:

* Request line structure
* HTTP method formatting
* HTTP protocol
* HTTP version
* Header names
* `Content-Length` values

The current implementation supports HTTP/1.1 requests.

# Header Handling

Headers are represented using a custom map type:

```go
type Headers map[string]string
```

Header names are normalized to lowercase internally, allowing case-insensitive lookup as required by HTTP.

The package provides methods for:

```go
Get()
Set()
Replace()
Delete()
Parse()
```

Multiple values assigned using `Set()` are combined using comma separation.

Header names are also validated to reject invalid characters.

# Response Writing

Responses are manually constructed and written through an `io.Writer`.

The response writer is responsible for writing:

1. The HTTP status line
2. Response headers
3. The response body
4. Chunked response bodies
5. HTTP trailers

Example HTTP response:

```text
HTTP/1.1 200 OK
content-type: text/html
content-length: 123
connection: close

<html>
    ...
</html>
```

The response writer supports the following status codes:

```text
200 OK
400 Bad Request
500 Internal Server Error
```

# Chunked Transfer Encoding

The project implements HTTP/1.1 chunked transfer encoding.

When the response size is not sent using `Content-Length`, the response body can be transmitted in chunks.

Each chunk follows this format:

```text
<chunk size in hexadecimal>\r\n
<chunk data>\r\n
```

For example:

```text
5\r\n
Hello\r\n
```

The response is terminated with a zero-length chunk:

```text
0\r\n
```

The implementation provides:

```go
WriteChunkedBody(p []byte)
WriteChunkedBodyDone()
```

The chunk size is encoded as a hexadecimal value before the chunk data is written.

# HTTP Trailers

The server also supports HTTP trailers.

When proxying content from `httpbingo.org`, the server calculates metadata about the complete response body and sends it after the chunked response.

The following trailers are generated:

```text
X-Content-SHA256
X-Content-Length
```

The SHA-256 trailer is generated using:

```go
crypto/sha256
```

The hash is encoded as a hexadecimal string using:

```go
encoding/hex
```

Example trailer headers:

```text
X-Content-SHA256: <sha256 hash>
X-Content-Length: <response length>
```

# Routes

## `/`

Returns a successful HTML response.

```text
HTTP/1.1 200 OK
```

---

## `/yourproblem`

Returns a `400 Bad Request` response.

```text
HTTP/1.1 400 Bad Request
```

---

## `/myproblem`

Returns a `500 Internal Server Error` response.

```text
HTTP/1.1 500 Internal Server Error
```

---

## `/httpbin/*`

Acts as a simple proxy for endpoints provided by `httpbingo.org`.

For example:

```text
/httpbin/get
```

is proxied to:

```text
https://httpbingo.org/get
```

The response is streamed back to the client using:

```text
Transfer-Encoding: chunked
```

The response also includes trailers containing:

* SHA-256 hash of the complete response
* Total response length

---

## `/video`

Serves an MP4 file from:

```text
assets/vim.mp4
```

The response uses:

```text
Content-Type: video/mp4
```

# Running the Server

Clone the repository:

```bash
git clone <https://github.com/ecetinerdem/httpfromtcp>
cd httpfromtcp
```

Run the server:

```bash
go run ./cmd/httpserver
```

The server starts on:

```text
localhost:42069
```

You should see:

```text
Server started on port 42069
```

# Testing the Server

You can test the server using `curl`.

## Root endpoint

```bash
curl http://localhost:42069/
```

## Bad request endpoint

```bash
curl http://localhost:42069/yourproblem
```

## Internal server error endpoint

```bash
curl http://localhost:42069/myproblem
```

## Proxy endpoint

```bash
curl http://localhost:42069/httpbin/get
```

## Streaming endpoint

You can test a streaming endpoint from `httpbingo.org`:

```bash
curl http://localhost:42069/httpbin/stream/10
```

To inspect headers and chunked transfer information:

```bash
curl -i http://localhost:42069/httpbin/get
```

# Graceful Shutdown

The server listens for the following operating system signals:

```text
SIGINT
SIGTERM
```

This allows the server to shut down gracefully when interrupted.

For example:

```bash
Ctrl + C
```

The server will log:

```text
Server gracefully stopped
```

# Technologies Used

* Go
* TCP sockets
* `net.Listener`
* `io.Reader`
* `io.Writer`
* HTTP/1.1
* Chunked Transfer Encoding
* HTTP Trailers
* SHA-256
* RFC-based HTTP concepts

# Learning Goals

The primary purpose of this project was to understand what normally happens behind Go's high-level `net/http` package.

Instead of using an existing HTTP server implementation, this project manually handles the lower-level components involved in HTTP communication.

Key concepts explored include:

* TCP connections are streams rather than messages
* A single `Read()` does not necessarily contain an entire HTTP request
* HTTP requests must be parsed incrementally
* HTTP uses `\r\n` delimiters
* Headers and bodies have different parsing rules
* `Content-Length` determines the expected size of a request body
* HTTP headers are case-insensitive
* HTTP responses are structured manually
* HTTP chunked transfer encoding uses hexadecimal chunk sizes
* HTTP trailers are sent after a chunked response
* `io.Reader` and `io.Writer` provide flexible abstractions for data streams

# Future Improvements

Possible future improvements include:

* Support additional HTTP methods
* Support more HTTP status codes
* Preserve multiple header values more accurately
* Improve error responses for malformed requests
* Add proper request routing
* Add concurrent connection handling if not already handled by the server implementation
* Stream large files instead of loading them entirely into memory
* Support persistent HTTP/1.1 connections
* Implement `Transfer-Encoding: chunked` request parsing
* Add automated tests for request parsing
* Add automated tests for response generation
* Add integration tests using raw TCP clients
* Improve RFC compliance

# Project Purpose

This project was created as a practical exercise in understanding HTTP by building the protocol components directly on top of TCP.

The goal was not to replace Go's production-ready `net/http` package, but to understand the abstractions that frameworks and HTTP libraries normally provide.

By manually implementing request parsing, header handling, response generation, chunked transfer encoding, and trailers, this project provides a lower-level view of how HTTP/1.1 communication works over a TCP connection.

