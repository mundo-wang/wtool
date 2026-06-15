package whttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
)

// WithBaseURL 设置请求的基础URL地址
func (cli *httpClient[T]) WithBaseURL(baseURL string) HttpClient[T] {
	cli.baseURL = baseURL
	return cli
}

// WithTimeout 设置请求超时时间，timeout为0表示不限制超时
func (cli *httpClient[T]) WithTimeout(timeout time.Duration) HttpClient[T] {
	if timeout > 0 {
		cli.client.Timeout = timeout
	}
	return cli
}

// WithRetry 配置请求重试策略，包括最大重试次数、首次重试延迟和最大重试延迟上限
func (cli *httpClient[T]) WithRetry(retryCount int, retryDelay, maxRetryDelay time.Duration) HttpClient[T] {
	if retryCount <= 0 {
		retryCount = 1
	}
	if retryDelay <= 0 {
		retryDelay = time.Second // 默认等待时间1秒
	}
	if maxRetryDelay <= 0 {
		maxRetryDelay = retryDelay * 16 // 给个默认最大值
	}
	cli.retryCount = retryCount
	cli.retryDelay = retryDelay
	cli.maxRetryDelay = maxRetryDelay
	return cli
}

// WithJsonBody 将传入的对象序列化为JSON并设置为请求体，同时自动添加Content-Type请求头
func (cli *httpClient[T]) WithJsonBody(body interface{}) HttpClient[T] {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		cli.err = err
		return cli
	}
	cli.jsonBody = jsonBody
	cli.WithHeader("Content-Type", "application/json")
	return cli
}

// 用于匹配baseURL模板中的占位符，如/user/{uid}/order/{oid}
var rePathVar = regexp.MustCompile(`\{([^{}]+)}`)

// WithPathParam 按顺序替换baseURL中的路径占位符（如{uid}），参数数量必须与占位符数量一致
func (cli *httpClient[T]) WithPathParam(args ...string) HttpClient[T] {
	matches := rePathVar.FindAllString(cli.baseURL, -1)
	if len(matches) != len(args) {
		err := fmt.Errorf("path param count mismatch: expected %d, got %d", len(matches), len(args))
		cli.err = err
		return cli
	}
	urlWithParams := cli.baseURL
	// 逐个替换占位符，参数值会进行url.PathEscape以保证路径安全
	for i, match := range matches {
		escaped := url.PathEscape(args[i])
		urlWithParams = strings.Replace(urlWithParams, match, escaped, 1)
	}
	cli.baseURL = urlWithParams
	return cli
}

// WithQueryParam 添加单个URL查询参数，value为空字符串时忽略
func (cli *httpClient[T]) WithQueryParam(key, value string) HttpClient[T] {
	if value != "" {
		cli.queryParams.Set(key, value)
	}
	return cli
}

// WithQueryParamByMap 通过键值对Map批量添加URL查询参数
func (cli *httpClient[T]) WithQueryParamByMap(params map[string]string) HttpClient[T] {
	for key, value := range params {
		cli.WithQueryParam(key, value)
	}
	return cli
}

// WithQueryParamByStruct 通过带url标签的结构体批量添加URL查询参数
func (cli *httpClient[T]) WithQueryParamByStruct(params interface{}) HttpClient[T] {
	queryParams, err := query.Values(params)
	if err != nil {
		cli.err = err
		return cli
	}
	for key, values := range queryParams {
		for _, value := range values {
			cli.WithQueryParam(key, value)
		}
	}
	return cli
}

// WithHeader 添加单个请求头，value为空字符串时忽略
func (cli *httpClient[T]) WithHeader(key, value string) HttpClient[T] {
	if value != "" {
		cli.headers[key] = value
	}
	return cli
}

// WithHeaderByMap 通过键值对Map批量添加请求头
func (cli *httpClient[T]) WithHeaderByMap(headers map[string]string) HttpClient[T] {
	for key, value := range headers {
		cli.WithHeader(key, value)
	}
	return cli
}

// Send 发送HTTP请求并将响应体反序列化为泛型类型T，失败时返回错误
func (cli *httpClient[T]) Send() (ResponseWrapper[T], error) {
	if cli.err != nil {
		return nil, cli.err
	}
	httpResp, err := cli.executeRequest()
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	return cli.handleResponse(httpResp)
}

// executeRequest 执行HTTP请求，支持带指数退避和随机抖动的超时重试机制
func (cli *httpClient[T]) executeRequest() (*http.Response, error) {
	var lastErr error
	attempts := 1 + cli.retryCount // 1次正常请求 + N次重试
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			// 指数退避：delay = retryDelay * 2^(attempt-1)
			delay := cli.retryDelay << (attempt - 1)
			if delay > cli.maxRetryDelay {
				delay = cli.maxRetryDelay
			}
			// 随机抖动：0.5 ~ 1.5 倍
			jitterFactor := 0.5 + rand.Float64() // rand.Float64() ∈ [0,1)
			jitterDelay := time.Duration(float64(delay) * jitterFactor)
			time.Sleep(jitterDelay)
		}
		// 每次创建新的Request，因为调用Do方法会导致Body内部数据被消耗
		req, err := cli.buildRequest()
		if err != nil {
			return nil, err
		}
		resp, err := cli.client.Do(req)
		if err == nil {
			return resp, nil
		}
		// 只对超时错误进行重试处理，非超时错误直接返回
		if !isTimeoutError(err) {
			return nil, err
		}
		lastErr = err
	}
	return nil, lastErr
}

// buildRequest 根据已配置的URL、查询参数、请求体和请求头构建http.Request对象
func (cli *httpClient[T]) buildRequest() (*http.Request, error) {
	var fullURL string
	if len(cli.queryParams) > 0 {
		fullURL = fmt.Sprintf("%s?%s", cli.baseURL, cli.queryParams.Encode())
	} else {
		fullURL = cli.baseURL
	}
	cli.fullURL = fullURL
	var body io.Reader
	if cli.jsonBody != nil {
		body = bytes.NewBuffer(cli.jsonBody)
	}
	req, err := http.NewRequest(cli.method, fullURL, body)
	if err != nil {
		return nil, err
	}
	for key, value := range cli.headers {
		req.Header.Set(key, value)
	}
	return req, nil
}

// handleResponse 读取HTTP响应体，2xx状态码时反序列化为T类型，否则返回包含状态码的错误
func (cli *httpClient[T]) handleResponse(resp *http.Response) (ResponseWrapper[T], error) {
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var respData T
		switch any(respData).(type) {
		// 如果T为[]byte或json.RawMessage，说明不需要反序列化JSON到具体类型，直接赋值字节数组
		case []byte:
			respData = any(respBytes).(T)
		case json.RawMessage:
			respData = any(json.RawMessage(respBytes)).(T)
		default:
			if len(respBytes) > 0 {
				if err = json.Unmarshal(respBytes, &respData); err != nil {
					return nil, err
				}
			}
		}
		handler := &responseWrapper[T]{
			respHeaders: resp.Header,
			respBytes:   respBytes,
			respData:    respData,
		}
		return handler, nil
	}
	err = fmt.Errorf("http status code not 2xx, is %d", resp.StatusCode)
	var errorResp map[string]any
	if jsonErr := json.Unmarshal(respBytes, &errorResp); jsonErr == nil {
		err = fmt.Errorf("%w, body: %v", err, errorResp)
	}
	return nil, err
}

// GetRespBytes 返回响应体的原始字节数组
func (cli *responseWrapper[T]) GetRespBytes() []byte {
	return cli.respBytes
}

// GetRespData 返回响应体反序列化后的泛型类型T对象
func (cli *responseWrapper[T]) GetRespData() T {
	return cli.respData
}

// GetRespHeader 获取响应头中指定key的第一个值，不存在则返回空字符串
func (cli *responseWrapper[T]) GetRespHeader(key string) string {
	value := cli.respHeaders.Get(key)
	return value
}

// GetRespHeaderMulti 获取响应头中指定key的所有值，不存在则返回空切片
func (cli *responseWrapper[T]) GetRespHeaderMulti(key string) []string {
	values := cli.respHeaders.Values(key)
	return values
}