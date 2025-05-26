package whttp

import (
	"net/http"
	"net/url"
	"time"
)

// 定义了httpClient所有需要暴露的方法
type HttpClient[T any] interface {
	WithJsonBody(body interface{}) HttpClient[T]
	WithPathParam(args ...string) HttpClient[T]
	WithQueryParam(key, value string) HttpClient[T]
	WithQueryParamByMap(params map[string]string) HttpClient[T]
	WithQueryParamByStruct(params interface{}) HttpClient[T]
	WithHeader(key, value string) HttpClient[T]
	WithHeaderByMap(headers map[string]string) HttpClient[T]
	Send() (HttpClient[T], error)
	GetResp() *T
	GetRespHeader(key string) string
	GetRespHeaderMulti(key string) []string
}

// HttpClient接口的实现结构体，私有
type httpClient[T any] struct {
	baseURL     string
	method      string
	queryParams url.Values
	jsonBody    []byte
	headers     map[string]string
	client      *http.Client
	err         error
	respHeaders http.Header
	respBytes   []byte
	resp        *T
}

// 泛型类型参数T表示返回的数据结构类型
func NewHttpClient[T any](baseURL, method string, timeout time.Duration) HttpClient[T] {
	transport := &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	}
	client := &http.Client{
		Transport: transport,
	}
	if timeout > 0 { // timeout = 0，代表不设置超时时间
		client.Timeout = timeout
	}
	return &httpClient[T]{
		baseURL:     baseURL,
		method:      method,
		queryParams: url.Values{},
		headers:     make(map[string]string),
		client:      client,
	}
}
