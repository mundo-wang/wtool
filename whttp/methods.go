package whttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/google/go-querystring/query"
	"github.com/mundo-wang/wtool/wlog" // 替换为指定的wlog路径
)

// 设置JSON请求体
func (cli *httpClient[T]) WithJsonBody(body interface{}) HttpClient[T] {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		wlog.Error("call json.Marshal failed").Err(err).Field("url", cli.baseURL).Field("method", cli.method).Log()
		cli.err = err
		return cli
	}
	cli.jsonBody = jsonBody
	cli.WithHeader("Content-Type", "application/json")
	return cli
}

// 设置路径参数
func (cli *httpClient[T]) WithPathParam(args ...string) HttpClient[T] {
	u, err := url.Parse(cli.baseURL)
	if err != nil {
		wlog.Error("call url.Parse failed").Err(err).Field("url", cli.baseURL).Field("method", cli.method).Log()
		cli.err = err
		return cli
	}
	segments := append([]string{u.Path}, args...)
	u.Path = path.Join(segments...)
	cli.baseURL = u.String()
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
		wlog.Error("call query.Values failed").Err(err).Field("url", cli.baseURL).Field("method", cli.method).Log()
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

// 设置请求头
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
func (cli *httpClient[T]) Send() HttpClient[T] {
	var fullURL string
	if len(cli.queryParams) > 0 {
		fullURL = fmt.Sprintf("%s?%s", cli.baseURL, cli.queryParams.Encode())
	} else {
		fullURL = cli.baseURL
	}
	var body io.Reader
	if cli.jsonBody != nil {
		body = bytes.NewBuffer(cli.jsonBody)
	}
	httpReq, err := http.NewRequest(cli.method, fullURL, body)
	if err != nil {
		wlog.Error("call http.NewRequest failed").Err(err).Field("url", fullURL).Field("method", cli.method).Log()
		cli.err = err
		return cli
	}
	for key, value := range cli.headers {
		httpReq.Header.Set(key, value)
	}
	httpResp, err := cli.client.Do(httpReq)
	if err != nil {
		wlog.Error("call cli.client.Do failed").Err(err).Field("url", fullURL).Field("method", cli.method).Log()
		cli.err = err
		return cli
	}
	defer httpResp.Body.Close()
	cli.respHeaders = httpResp.Header
	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		wlog.Error("call io.ReadAll failed").Err(err).Field("url", fullURL).Field("method", cli.method).Log()
		cli.err = err
		return cli
	}
	// 由于一些HTTP接口返回的成功状态码不一定为200，所以这里判断只要是2开头的状态码，均视为请求成功
	if httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 {
		cli.respBytes = respBytes
		resp := new(T)
		err = json.Unmarshal(respBytes, resp)
		if err != nil {
			wlog.Error("call json.Unmarshal failed").Err(err).Field("url", fullURL).Field("method", cli.method).Log()
			cli.err = err
		}
		cli.resp = resp
		return cli
	}
	var errorResp map[string]interface{}
	err = json.Unmarshal(respBytes, &errorResp)
	if err != nil {
		wlog.Error("call json.Unmarshal failed").Err(err).Field("url", fullURL).
			Field("method", cli.method).Field("statusCode", httpResp.StatusCode).Log()
		cli.err = err
		return cli
	}
	err = fmt.Errorf("status code not 200, is %d", httpResp.StatusCode)
	wlog.Error("call cli.client.Do failed").Err(err).Field("url", fullURL).
		Field("method", cli.method).Field("errorResp", errorResp).Log()
	cli.err = err
	return cli
}

// 检查发送请求过程中是否有报错
func (cli *httpClient[T]) Error() error {
	return cli.err
}

// 返回响应体的字节数组
func (cli *httpClient[T]) GetRespBytes() []byte {
	return cli.respBytes
}

// 返回响应体反序列化的对象
func (cli *httpClient[T]) GetResp() *T {
	return cli.resp
}

// GetRespHeader 获取指定key关联的第一个值，如果无关联，返回空字符串
func (cli *httpClient[T]) GetRespHeader(key string) string {
	value := cli.respHeaders.Get(key)
	return value
}

// GetRespHeaderMulti 获取指定key关联的所有值，如果无关联，返回空切片
func (cli *httpClient[T]) GetRespHeaderMulti(key string) []string {
	values := cli.respHeaders.Values(key)
	return values
}
