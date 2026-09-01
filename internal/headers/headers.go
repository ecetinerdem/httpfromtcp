package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(headerBytes []byte) (n int, done bool, err error) {

	idx := bytes.Index(headerBytes, []byte("\r\n"))

	if idx == -1 {
		return 0, false, nil
	}

	if idx == 0 {
		return 2, true, nil
	}

	parts := bytes.SplitN(headerBytes[:idx], []byte(":"), 2)

	header := string(parts[0])
	value := string(parts[1])
	if strings.HasPrefix(header, " ") {
		return 0, false, fmt.Errorf("invalid headers")
	}

	header = strings.TrimSpace(header)
	if !validHeaderName(header) {
		return 0, false, fmt.Errorf("invalid characters in headers")
	}
	value = strings.TrimSpace(value)
	if strings.Contains(value, "/") && !strings.Contains(value, "*/*") && strings.HasPrefix(value, "curl") {
		valueSlice := strings.SplitN(value, "/", 2)
		value = valueSlice[0]
		_ = valueSlice[1]
	}

	h.Set(header, value)

	return (len(headerBytes[:idx]) + len("\r\n")), false, nil

}

func (h Headers) Get(key string) (string, bool) {
	key = strings.ToLower(key)

	value, ok := h[key]
	return value, ok
}

func (h Headers) Set(key string, value string) {

	key = strings.ToLower(key)

	existingValue, ok := h[key]
	if !ok {
		h[key] = value
		return
	}

	h[key] = fmt.Sprintf("%s, %s", existingValue, value)

}

func (h Headers) Replace(key string, value string) {
	key = strings.ToLower(key)
	h[key] = value
}

func (h Headers) Delete(key string) {
	key = strings.ToLower(key)
	delete(h, key)
}

func validHeaderName(s string) bool {
	for _, r := range s {
		if r < 33 || r > 126 {
			return false
		}

		switch r {
		case '(', ')', '<', '>', '@', ',', ';', ':',
			'\\', '"', '/', '[', ']', '?', '=', '{', '}':
			return false
		}
	}

	return true
}
