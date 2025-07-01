package whttp

import (
	"net/http"
	"net/url"
	"time"
)

type HttpClient[T any] interface {
	WithBaseURL(baseURL string) HttpClient[T]
	WithTimeout(timeout time.Duration) HttpClient[T]
	WithJsonBody(body interface{}) HttpClient[T]
	WithPathParam(args ...string) HttpClient[T]
	WithQueryParam(key, value string) HttpClient[T]
	WithQueryParamByMap(params map[string]string) HttpClient[T]
	WithQueryParamByStruct(params interface{}) HttpClient[T]
	WithHeader(key, value string) HttpClient[T]
	WithHeaderByMap(headers map[string]string) HttpClient[T]
	Send() (ResponseHandler[T], error)
}

type httpClient[T any] struct {
	baseURL     string
	method      string
	fullURL     string
	queryParams url.Values
	jsonBody    []byte
	headers     map[string]string
	client      *http.Client
	err         error
}

type ResponseHandler[T any] interface {
	GetRespBytes() []byte
	GetRespData() T
	GetRespHeader(key string) string
	GetRespHeaderMulti(key string) []string
}

type responseHandler[T any] struct {
	respHeaders http.Header
	respBytes   []byte
	respData    T
}

func NewGet[T any]() HttpClient[T] {
	return newHttpClient[T](http.MethodGet)
}

func NewPost[T any]() HttpClient[T] {
	return newHttpClient[T](http.MethodPost)
}

func NewPut[T any]() HttpClient[T] {
	return newHttpClient[T](http.MethodPut)
}

func NewPatch[T any]() HttpClient[T] {
	return newHttpClient[T](http.MethodPatch)
}

func NewDelete[T any]() HttpClient[T] {
	return newHttpClient[T](http.MethodDelete)
}

// 泛型类型参数T表示返回的数据结构类型
func newHttpClient[T any](method string) HttpClient[T] {
	// 这里设置的都是默认值，可根据后端服务场景进行修改
	transport := &http.Transport{
		MaxIdleConns:        100,              // 全局最大空闲连接数
		MaxIdleConnsPerHost: 2,                // 每个目标主机最大空闲连接数
		IdleConnTimeout:     90 * time.Second, // 空闲连接最大存活时间，超过则关闭连接
	}
	client := &http.Client{
		Transport: transport,
	}
	return &httpClient[T]{
		method:      method,
		queryParams: url.Values{},
		headers:     make(map[string]string),
		client:      client,
	}
}
