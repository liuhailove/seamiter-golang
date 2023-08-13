package client

import (
	"errors"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	MaxBodySize = 1024 * 1024 * 4
)

// SimpleHttpResponseParser
// * The parser provides functionality to parse raw bytes HTTP response to a {@link SimpleHttpResponse}.
// * </p>
// * <p>
// * Note that this is a very NAIVE parser, {@code Content-Length} must be specified in the
// * HTTP response header, otherwise, the body will be dropped. All other body type such as
// * {@code Transfer-Encoding: chunked}, {@code Transfer-Encoding: deflate} are not supported.
type SimpleHttpResponseParser struct {
	buf []byte
}

func NewSimpleHttpResponseParserWithMaxSize(maxSize int) *SimpleHttpResponseParser {
	if maxSize < 0 {
		panic(errors.New("maxSize must >0"))
	}
	var parser = new(SimpleHttpResponseParser)
	parser.buf = make([]byte, maxSize)
	return parser
}

func NewSimpleHttpResponseParser() *SimpleHttpResponseParser {
	return NewSimpleHttpResponseParserWithMaxSize(1024 * 4)
}

func (s SimpleHttpResponseParser) Parse(resp *http.Response) *SimpleHttpResponse {
	var header = make(map[string]string)
	for key, values := range resp.Header {
		header[key] = strings.Join(values, ",")
	}
	var response = NewSimpleHttpResponse(resp.Status, resp.StatusCode, header)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logging.Warn("[SimpleHttpResponseParser] ReadAll err", "error msg", err)
		return response
	}
	response.SetBody(body)
	return response
}
