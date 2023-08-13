package client

import (
	"bytes"
	"errors"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/transport/common/transport/endpoint"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// SimpleHttpClient
//  * <p>
// * A very simple HTTP client that only supports GET/POST method and plain text request body.
// * The Content-Type header is always set as <pre>application/x-www-form-urlencoded</pre>.
// * All parameters in the request will be encoded using {@link URLEncoder#encode(String, String)}.
// * </p>
// * <p>
// * The result of a HTTP invocation will be wrapped as a {@link SimpleHttpResponse}. Content in response body
// * will be automatically decoded to string with provided charset.
// * </p>
// * <p>
// * This is a blocking and synchronous client, so an invocation will await the response until timeout exceed.
// * </p>
// * <p>
// * Note that this is a very NAIVE client, {@code Content-Length} must be specified in the
// * HTTP response header, otherwise, the response body will be dropped. All other body type such as
// * {@code Transfer-Encoding: chunked}, {@code Transfer-Encoding: deflate} are not supported.
// * </p>
type SimpleHttpClient struct {
	httpClient http.Client
}

func (s SimpleHttpClient) Initial(proxyUrl string) {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
		MaxConnsPerHost:       400,
		MaxIdleConnsPerHost:   100,
	}
	if len(strings.TrimSpace(proxyUrl)) != 0 {
		proxy, err := url.Parse(proxyUrl)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxy)
		} else {
			logging.Info("parser proxy error:" + err.Error())
		}
	}
	s.httpClient = http.Client{
		Timeout:   time.Duration(3) * time.Second,
		Transport: transport,
	}
}

// Get
// * Execute a GET HTTP request.
// *
// * @param request HTTP request
// * @return the response if the request is successful
// * @throws IOException when connection cannot be established or the connection is interrupted
func (s SimpleHttpClient) Get(req *SimpleHttpRequest) (*SimpleHttpResponse, error) {
	if req == nil {
		return nil, nil
	}
	return s.request(req.GetEndpoint(), http.MethodGet, req.GetRequestPath(), req.GetParams())
}

// Post
// * Execute a POST HTTP request.
// *
// * @param request HTTP request
// * @return the response if the request is successful
// * @throws IOException when connection cannot be established or the connection is interrupted
func (s SimpleHttpClient) Post(req *SimpleHttpRequest) (*SimpleHttpResponse, error) {
	if req == nil {
		return nil, errors.New("[SimpleHttpClient] req cannot be nil")
	}
	return s.request(req.GetEndpoint(), http.MethodPost, req.GetRequestPath(), req.GetParams())
}

func (s SimpleHttpClient) request(epoint endpoint.Endpoint, methodType string, requestPath string, paramsMap map[string]string) (*SimpleHttpResponse, error) {
	var httpClient = s.httpClient
	var url = "http://"
	if epoint.Protocol == endpoint.HTTPS {
		url = "https://"
	}
	url += epoint.Host
	if epoint.Port > 0 {
		url += ":" + strconv.Itoa(int(epoint.Port))
	}
	requestPath = s.getRequestPath(methodType, requestPath, paramsMap, "")
	requestPath = url + requestPath
	var req *http.Request
	var err error
	if methodType == http.MethodGet {
		req, err = http.NewRequest(http.MethodGet, requestPath, nil)
	} else {
		// POST method
		req, err = http.NewRequest(http.MethodPost, requestPath, bytes.NewBufferString(s.encodeRequestParams(paramsMap, "")))
	}
	if err != nil {
		logging.Warn("[SimpleHttpClient] request error", "msg", err)
		return nil, err
	}
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		logging.Warn("[SimpleHttpClient] request do error", "msg", err)
		return nil, err
	}
	defer resp.Body.Close()
	response := NewSimpleHttpResponseParser().Parse(resp)
	return response, nil
}

// GetRequestPath
// 获取请求路径
func (s SimpleHttpClient) getRequestPath(methodType string, requestPath string, paramsMap map[string]string, charset string) string {
	if methodType == http.MethodGet {
		if strings.Contains(requestPath, "?") {
			return requestPath + "&" + s.encodeRequestParams(paramsMap, charset)
		}
		return requestPath + "?" + s.encodeRequestParams(paramsMap, charset)
	}
	return requestPath
}

// encodeRequestParams
// * Encode and get the URL request parameters.
// *
// * @param paramsMap pair of parameters
// * @param charset   charset
//  * @return encoded request parameters, or empty string ("") if no parameters are provided
func (s SimpleHttpClient) encodeRequestParams(paramsMap map[string]string, charset string) string {
	if paramsMap == nil || len(paramsMap) == 0 {
		return ""
	}
	var paramsBuilder bytes.Buffer
	for k, v := range paramsMap {
		if k == "" || v == "" {
			continue
		}
		paramsBuilder.WriteString(k)
		paramsBuilder.WriteString("=")
		paramsBuilder.WriteString(v)
		paramsBuilder.WriteString("&")
	}

	if paramsBuilder.Len() > 0 {
		paramsBuilder.Truncate(paramsBuilder.Len() - 1)
	}
	return paramsBuilder.String()
}
