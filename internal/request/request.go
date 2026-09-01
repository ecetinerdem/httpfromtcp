package request

import (
	"bytes"
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
	"strings"
)

const (
	initialized                = "initialized"
	requestStateParsingHeaders = "requestStateParsingHeaders"
	requestStateParsingBody    = "requestStateParsingBody"
	done                       = "done"
)

const (
	bufferSize = 8
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	State       string
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) parse(data []byte) (int, error) {

	totalBytesParsed := 0
	for r.State != done {

		switch r.State {
		case initialized:
			method, target, version, nBytes, err := parseRequestLine(data)
			if err != nil {
				return 0, err
			}

			if nBytes == 0 {
				return 0, nil
			}

			totalBytesParsed += nBytes
			r.RequestLine.HttpVersion = version
			r.RequestLine.Method = method
			r.RequestLine.RequestTarget = target
			r.State = requestStateParsingHeaders
		case requestStateParsingHeaders:
			nBytes, doneParse, err := r.Headers.Parse(data[totalBytesParsed:])

			if err != nil {
				if errors.Is(err, io.EOF) {
					if r.State != done {
						return 0, fmt.Errorf("incomplete request, in state %s...", requestStateParsingHeaders)

					}
					break
				}
				return 0, err
			}

			if nBytes == 0 {
				return totalBytesParsed, nil
			}

			if !doneParse {
				totalBytesParsed += nBytes
			} else {
				totalBytesParsed += nBytes
				r.State = requestStateParsingBody
			}
		case requestStateParsingBody:
			value, ok := r.Headers.Get("Content-Length")

			if !ok {

				remainingBytes := len(data[totalBytesParsed:])
				totalBytesParsed += remainingBytes
				r.State = done

				return totalBytesParsed, nil
			}
			valueInt, err := strconv.Atoi(value)
			if err != nil {
				return 0, fmt.Errorf("incorrect request header information in state %s", requestStateParsingBody)

			}

			remainingBytes := len(data[totalBytesParsed:])

			r.Body = append(r.Body, data[totalBytesParsed:totalBytesParsed+remainingBytes]...)
			totalBytesParsed += remainingBytes

			if len(r.Body) > valueInt {
				return 0, fmt.Errorf("incorrect request header information in state %s", requestStateParsingBody)
			}

			if valueInt == len(r.Body) {
				r.State = done
				return totalBytesParsed, nil
			}

			fmt.Printf("%d bytes consumed\n", totalBytesParsed)
			return totalBytesParsed, nil
		case done:
			return 0, fmt.Errorf("error: trying to read data in a done state")
		default:
			return 0, fmt.Errorf("error: unknown state")
		}
	}
	return totalBytesParsed, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {

	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0
	var request Request
	request.State = initialized
	request.Headers = headers.NewHeaders()
	for request.State != done {

		if readToIndex == len(buf) {
			buf = append(buf, make([]byte, len(buf))...)
		}
		nBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if request.State != done {
					return nil, fmt.Errorf("incomplete request, in state %s, read n bytes on EOF: %d", request.State, readToIndex)
				}
				break
			}
			return nil, err
		}

		readToIndex += nBytesRead
		n, err := request.parse(buf[:readToIndex])

		if err != nil {
			return nil, err
		}

		copy(buf, buf[n:readToIndex])
		readToIndex -= n

	}

	return &request, nil

}

func parseRequestLine(requestBytes []byte) (string, string, string, int, error) {
	idx := bytes.Index(requestBytes, []byte("\r\n"))
	var line string
	if idx != -1 {
		line = string(requestBytes[:idx])
	} else {
		return "", "", "", 0, nil
	}

	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return "", "", "", len(requestBytes[:idx]) + len("\r\n"), fmt.Errorf("malformed requestline")
	}

	if !isUpper(parts[0]) {
		return "", "", "", len(requestBytes[:idx]) + len("\r\n"), fmt.Errorf("malformed method in requestline")
	}

	versionLine := parts[2]
	versionParts := strings.Split(versionLine, "/")
	if len(versionParts) != 2 {
		return "", "", "", len(requestBytes[:idx]) + len("\r\n"), fmt.Errorf("malformed version line")
	}
	versionProtocol := versionParts[0]
	versionDigit := versionParts[1]

	if ("1.1" != versionDigit) || ("HTTP" != versionProtocol) {
		return "", "", "", len(requestBytes[:idx]) + len("\r\n"), fmt.Errorf("wrong or malformed version or protocol in requestline")
	}
	method := parts[0]
	target := parts[1]
	version := versionDigit

	return method, target, version, len(requestBytes[:idx]) + len("\r\n"), nil

}

func isUpper(s string) bool {
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}
