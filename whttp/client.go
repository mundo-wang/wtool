package whttp

import (
	"net/http"
	"net/url"
	"time"
)

// HttpClient 定义了httpClient所有需要暴露的方法
type HttpClient interface {
	WithJsonBody(body interface{}) HttpClient
	WithPathParam(args ...string) HttpClient
	WithQueryParam(key, value string) HttpClient
	WithQueryParamByMap(params map[string]string) HttpClient
	WithQueryParamByStruct(params interface{}) HttpClient
	WithHeader(key, value string) HttpClient
	WithHeaderByMap(headers map[string]string) HttpClient
	Send() ([]byte, error)
	GetRespHeader(key string) string
	GetRespHeaders(key string) []string
}

// httpClient HttpClient接口的实现结构体，私有
type httpClient struct {
	baseURL     string
	method      string
	queryParams url.Values
	jsonBody    []byte
	headers     map[string]string
	client      *http.Client
	respHeaders http.Header
	err         error
}

// NewHttpClient 创建一个新的httpClient实例
func NewHttpClient(baseURL, method string, timeout time.Duration) HttpClient {
	transport := &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	}
	client := &http.Client{
		Transport: transport,
	}
	if timeout > 0 {
		client.Timeout = timeout
	}
	return &httpClient{
		baseURL:     baseURL,
		method:      method,
		queryParams: url.Values{},
		headers:     make(map[string]string),
		client:      client,
	}
}
