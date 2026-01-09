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

func (cli *httpClient[T]) WithBaseURL(baseURL string) HttpClient[T] {
	cli.baseURL = baseURL
	return cli
}

// 设置接口超时时间，如果timeout = 0，代表无超时时间
func (cli *httpClient[T]) WithTimeout(timeout time.Duration) HttpClient[T] {
	if timeout > 0 {
		cli.client.Timeout = timeout
	}
	return cli
}

// retryCount为最大重试次数，retryDelay为基础等待时间，maxRetryDelay为最大等待时间
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

// 设置JSON请求体
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

// 填充baseURL中的路径参数，支持占位符替换
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

// 设置单个查询参数
func (cli *httpClient[T]) WithQueryParam(key, value string) HttpClient[T] {
	if value != "" {
		cli.queryParams.Set(key, value)
	}
	return cli
}

// 通过map设置多个查询参数
func (cli *httpClient[T]) WithQueryParamByMap(params map[string]string) HttpClient[T] {
	for key, value := range params {
		cli.WithQueryParam(key, value)
	}
	return cli
}

// 通过结构体对象，设置多个查询参数
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

// 设置单个请求头
func (cli *httpClient[T]) WithHeader(key, value string) HttpClient[T] {
	if value != "" {
		cli.headers[key] = value
	}
	return cli
}

// 通过map设置多个请求头
func (cli *httpClient[T]) WithHeaderByMap(headers map[string]string) HttpClient[T] {
	for key, value := range headers {
		cli.WithHeader(key, value)
	}
	return cli
}

// 发送请求并封装响应
func (cli *httpClient[T]) Send() (ResponseWrapper[T], error) {
	if cli.err != nil {
		return nil, cli.err
	}
	httpReq, err := cli.buildRequest()
	if err != nil {
		return nil, err
	}
	httpResp, err := cli.executeRequest(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	return cli.handleResponse(httpResp)
}

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

func (cli *httpClient[T]) executeRequest(req *http.Request) (*http.Response, error) {
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

func (cli *httpClient[T]) handleResponse(resp *http.Response) (ResponseWrapper[T], error) {
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var respData T
		if err = json.Unmarshal(respBytes, &respData); err != nil {
			return nil, err
		}
		handler := &responseWrapper[T]{
			respHeaders: resp.Header,
			respBytes:   respBytes,
			respData:    respData,
		}
		return handler, nil
	}
	var errorResp map[string]any
	if err = json.Unmarshal(respBytes, &errorResp); err != nil {
		return nil, err
	}
	err = fmt.Errorf("http status code not 200, is %d", resp.StatusCode)
	return nil, err
}

// 返回响应体的字节数组
func (cli *responseWrapper[T]) GetRespBytes() []byte {
	return cli.respBytes
}

// 返回响应体反序列化的对象
func (cli *responseWrapper[T]) GetRespData() T {
	return cli.respData
}

// GetRespHeader 获取指定key关联的第一个值，如果无关联，返回空字符串
func (cli *responseWrapper[T]) GetRespHeader(key string) string {
	value := cli.respHeaders.Get(key)
	return value
}

// GetRespHeaderMulti 获取指定key关联的所有值，如果无关联，返回空切片
func (cli *responseWrapper[T]) GetRespHeaderMulti(key string) []string {
	values := cli.respHeaders.Values(key)
	return values
}
