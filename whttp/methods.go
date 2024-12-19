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

// WithJsonBody 设置JSON请求体
func (cli *httpClient) WithJsonBody(body interface{}) HttpClient {
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

// WithPathParam 设置路径参数
func (cli *httpClient) WithPathParam(args ...string) HttpClient {
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

// WithQueryParam 设置单个查询参数
func (cli *httpClient) WithQueryParam(key, value string) HttpClient {
	if value != "" {
		cli.queryParams.Set(key, value)
	}
	return cli
}

// WithQueryParamByMap 设置查询参数（通过map）
func (cli *httpClient) WithQueryParamByMap(params map[string]string) HttpClient {
	for key, value := range params {
		cli.WithQueryParam(key, value)
	}
	return cli
}

// WithQueryParamByStruct 设置查询参数（通过结构体）
func (cli *httpClient) WithQueryParamByStruct(params interface{}) HttpClient {
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

// WithHeader 设置请求头
func (cli *httpClient) WithHeader(key, value string) HttpClient {
	if value != "" {
		cli.headers[key] = value
	}
	return cli
}

// WithHeaderByMap 设置请求头（通过map）
func (cli *httpClient) WithHeaderByMap(headers map[string]string) HttpClient {
	for key, value := range headers {
		cli.WithHeader(key, value)
	}
	return cli
}

// Send 发送请求并返回响应
func (cli *httpClient) Send() ([]byte, error) {
	if cli.err != nil {
		return nil, cli.err
	}
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
		return nil, err
	}
	for key, value := range cli.headers {
		httpReq.Header.Set(key, value)
	}
	httpResp, err := cli.client.Do(httpReq)
	if err != nil {
		wlog.Error("call cli.client.Do failed").Err(err).Field("url", fullURL).Field("method", cli.method).Log()
		return nil, err
	}
	defer httpResp.Body.Close()
	cli.respHeaders = httpResp.Header
	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		wlog.Error("call io.ReadAll failed").Err(err).Field("url", fullURL).Field("method", cli.method).Log()
		return nil, err
	}
	// 由于一些HTTP接口返回的成功状态码不一定为200，所以这里判断只要是2开头的状态码，均视为请求成功
	if httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 {
		return respBytes, nil
	}
	var errorResp map[string]interface{}
	err = json.Unmarshal(respBytes, &errorResp)
	if err != nil {
		wlog.Error("call json.Unmarshal failed").Err(err).Field("url", fullURL).
			Field("method", cli.method).Field("statusCode", httpResp.StatusCode).Log()
		return nil, err
	}
	err = fmt.Errorf("status code not 200, is %d", httpResp.StatusCode)
	wlog.Error("call cli.client.Do failed").Err(err).Field("url", fullURL).
		Field("method", cli.method).Field("errorResp", errorResp).Log()
	return nil, err
}

// GetRespHeader 获取指定key关联的第一个值，如果无关联，返回空字符串
func (cli *httpClient) GetRespHeader(key string) string {
	value := cli.respHeaders.Get(key)
	return value
}

// GetRespHeaderMulti 获取指定key关联的所有值，如果无关联，返回空切片
func (cli *httpClient) GetRespHeaderMulti(key string) []string {
	values := cli.respHeaders.Values(key)
	return values
}
