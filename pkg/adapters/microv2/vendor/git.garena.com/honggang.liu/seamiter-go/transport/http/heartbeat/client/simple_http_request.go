package client

import (
	"errors"
	"git.garena.com/honggang.liu/seamiter-go/transport/common/transport/endpoint"
)

type SimpleHttpRequest struct {
	endpoint    endpoint.Endpoint
	requestPath string
	soTimeout   int64
	params      map[string]string
}

func NewSimpleHttpRequest(endpoint endpoint.Endpoint, requestPath string) *SimpleHttpRequest {
	req := new(SimpleHttpRequest)
	req.endpoint = endpoint
	req.requestPath = requestPath
	req.soTimeout = 3000
	return req
}

func (s *SimpleHttpRequest) GetEndpoint() endpoint.Endpoint {
	return s.endpoint
}

func (s *SimpleHttpRequest) SetEndpoint(endpoint endpoint.Endpoint) {
	s.endpoint = endpoint
}

func (s *SimpleHttpRequest) SetRequestPath(requestPath string) {
	s.requestPath = requestPath
}

func (s *SimpleHttpRequest) GetSoTimeout() int64 {
	return s.soTimeout
}

func (s *SimpleHttpRequest) SetSoTimeout(soTimeout int64) {
	s.soTimeout = soTimeout
}

func (s *SimpleHttpRequest) GetParams() map[string]string {
	return s.params
}

func (s *SimpleHttpRequest) SetParams(params map[string]string) {
	s.params = params
}

func (s *SimpleHttpRequest) AddParam(key, value string) {
	if key == "" {
		panic(errors.New("parameter key cannot be empty"))
	}
	if s.params == nil {
		s.params = make(map[string]string)
	}
	s.params[key] = value
}

func (s *SimpleHttpRequest) GetRequestPath() string {
	return s.requestPath
}
