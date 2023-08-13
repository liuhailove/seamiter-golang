package client

import (
	"bytes"
	"strconv"
	"strings"
)

type SimpleHttpResponse struct {
	statusLine string
	statusCode int
	headers    map[string]string
	body       []byte
	charset    string
}

func NewSimpleHttpResponse(statusLine string, statusCode int, headers map[string]string) *SimpleHttpResponse {
	res := new(SimpleHttpResponse)
	res.statusLine = statusLine
	res.statusCode = statusCode
	res.headers = headers
	return res
}

func NewSimpleHttpResponseWithBody(statusLine string, headers map[string]string, body []byte) *SimpleHttpResponse {
	res := new(SimpleHttpResponse)
	res.statusLine = statusLine
	res.headers = headers
	res.body = body
	return res
}

func (s *SimpleHttpResponse) parserCharset() {
	contentType := s.GetHeader("Content-Type")
	for _, str := range strings.Split(contentType, " ") {
		if strings.HasPrefix(strings.ToLower(str), "charset=") {
			s.charset = strings.Split(str, "=")[1]
		}
	}
}

func (s *SimpleHttpResponse) parseCode() {
	s.statusCode, _ = strconv.Atoi(strings.Split(s.statusLine, " ")[1])
}

func (s *SimpleHttpResponse) SetBody(body []byte) {
	s.body = body
}

func (s *SimpleHttpResponse) GetBody() []byte {
	return s.body
}

func (s *SimpleHttpResponse) GetStatusLine() string {
	return s.statusLine
}

func (s *SimpleHttpResponse) GetStatusCode() int {
	if s.statusCode == 0 {
		s.parseCode()
	}
	return s.statusCode
}

func (s *SimpleHttpResponse) GetHeaders() map[string]string {
	return s.headers
}

func (s *SimpleHttpResponse) GetHeader(key string) string {
	if s.headers == nil {
		return ""
	}
	value := s.headers[key]
	if value != "" {
		return value
	}
	for k, v := range s.headers {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return ""
}

func (s *SimpleHttpResponse) GetBodyAsString() string {
	s.parserCharset()
	return string(s.body)
}

func (s *SimpleHttpResponse) String() string {
	var buf bytes.Buffer
	buf.WriteString(s.statusLine)
	buf.WriteString("\r\n")
	if s.headers != nil {
		for k, v := range s.headers {
			buf.WriteString(k)
			buf.WriteString(": ")
			buf.WriteString(v)
			buf.WriteString("\r\n")
		}
	}
	buf.WriteString("\r\n")
	buf.WriteString(s.GetBodyAsString())
	return buf.String()
}
