在开发过程中，我编写了一些便捷的小工具，这些工具经过了实战检验，功能稳定且实用。现在，我对它们进行了优化和整理，分享出来，大家可以直接在`Go`项目中使用。

可以通过以下命令将工具包引入项目：

```shell
go get github.com/mundo-wang/wtool
```

### 1. 日志工具

首先在需要的模块中导入`wlog`包，代码中通过链式调用来实现日志记录。代码示例如下：

```go
wlog.Warnf("Hello: %s", "gopher").Field("name", "zhangsan").Field("age", 30).Log()
```

打印在控制台的结果如下所示：

```sh
2024-04-22 15:47:03	WARN	prac/main.go:17	Hello: gopher	{"name": "zhangsan", "age": 30, "caller": "common.Hello"}
```

在`Goland`点一下`prac/main.go:17`的部分，可以直接跳转到代码中打印这条日志的地方，也可以复制这个代码位置信息全文查找。

如果是生产环境，打印的日志是这个样子的，这是一份标准的`JSON`格式数据：

```json
{"level":"WARN","time":"2024-04-22 15:46:22","line":"prac/main.go:17","message":"Hello: gopher","name":"zhangsan","age":30,"caller":"common.Hello"}
```

在这里，我们通过`os.Getenv("env")`来判断当前环境是否为生产环境。若使用`Docker`启动容器，只需添加`-e env=production`参数即可使日志进入生产环境模式。若通过执行可执行文件运行项目，在执行命令前运行`export env=production`即可启用生产环境模式。


如果需要在日志中打印`error`，代码示例如下：

```go
err := errors.New("some errors")
wlog.Error("call xxx failed").Err(err).Field("name", "lisi").Log()
```

打印出的生产环境`JSON`格式日志如下所示：

```json
{"level":"ERROR","time":"2024-12-16 09:43:08","line":"test06/main.go:46","message":"call xxx failed","error":"some errors","name":"lisi","caller":"main.CallSome"}
```

### 2. HTTP工具

我们在使用`http`库调用公共接口时，通常需要执行以下步骤：

1. 指定待访问的`URL`（对于`GET`请求，需要拼接参数进`URL`，对于`POST`请求，需要预备请求体`JSON`的`[]byte`对象）。
2. 创建`HTTP client`，设置自定义参数，例如接口的请求超时时长等。
3. 创建`httpReq`，指定请求方法、`URL`、请求体参数（若有），并在请求头中放置参数如`Content-Type`等。
4. 使用`client.Do(httpReq)`，调用接口请求，并获取到响应对象`httpResp`。
5. 处理`httpResp`，如判断其`StatusCode`属性是否为`200`，并用`io.ReadAll`从`Body`中获取到响应体内容。
6. 使用`json.Unmarshal`将响应体的`[]byte`数据反序列化为对应的结构体或`map`等对象。

这个过程步骤非常繁琐，需要记住整个步骤，还需要编写大量代码。`HTTP`工具采用链式调用，把这个过程串联起来。

下面以一个`Get`请求和一个`Post`请求为例，讲一下上面日志工具的用法：

- 服务器`IP:Port`：`10.40.18.34:8080`
- 请求`URL`：`http://10.40.18.34:8080/set_user`
- 请求方式：`GET`
- 请求参数：`username`、`age`（必选），`address`（可选）
- 请求头：`Authorization=a96902a7-bc99-6d2fb2bf1569`

使用我们的`HTTP`工具完成调用过程，代码示例如下：

```go
type User struct {
	Username string `url:"username"`
	Age      int    `url:"age"`
	Address  string `url:"address,omitempty"`
}

func main() {
	baseURL := "http://10.40.18.34:8080/set_user"
	user := &User{
		Username: "zhangsan",
		Age:      30,
		Address:  "蔡徐村",
	}
	respBytes, _ := whttp.NewHttpClient(baseURL, http.MethodGet, 10*time.Second).
		WithHeader("Authorization", "a96902a7-bc99-6d2fb2bf1569").WithQueryParamByStruct(user).Send()
	fmt.Println(string(respBytes))
}
```

我们也可以使用`WithQueryParam`方法继续往后面补充`query`参数：

```go
respBytes, _ := whttp.NewHttpClient(baseURL, http.MethodGet, 10*time.Second).
		WithHeader("Authorization", "a96902a7-bc99-6d2fb2bf1569").
		WithQueryParamByStruct(user).WithQueryParam("address", "caixucun").Send()
```

- 服务器`IP:Port`：`10.40.18.34:8080`
- 请求`URL`：`http://10.40.18.34:8080/set_book`
- 请求方式：`POST`
- 请求参数：`title`、`name`、`auther`（必选），`price`（可选）
- 请求头：`Authorization=a96902a7-bc99-6d2fb2bf1569`、`Content-Type=application/json`

使用我们的`HTTP`工具完成调用过程，代码示例如下：

```go
type Book struct {
	Title  string  `json:"title"`
	Name   string  `json:"name"`
	Author string  `json:"author"`
	Price  float64 `json:"price,omitempty"`
}

func main() {
	baseURL := "http://10.40.18.34:8080/set_book"
	book := &Book{
		Title:  "科技",
		Name:   "MySQL必知必会",
		Author: "Java之父余胜军",
		Price:  59.99,
	}
	respBytes, _ := whttp.NewHttpClient(baseURL, http.MethodPost, 5*time.Second).
		WithHeader("Authorization", "a96902a7-bc99-6d2fb2bf1569").WithJsonBody(book).Send()
	fmt.Println(string(respBytes))
}
```

如果想获取响应头中的指定参数，可以使用以下代码方式：

```go
httpClient := whttp.NewHttpClient(baseURL, http.MethodPost, 5*time.Second).
	WithHeader("Authorization", "a96902a7-bc99-6d2fb2bf1569").WithJsonBody(book)
respBytes, _ := httpClient.Send()
authToken := httpClient.GetRespHeader("authToken")
```

目前，该`HTTP`工具仅支持`POST`请求在请求体中使用`JSON`格式传递参数，对于表单或其他格式暂不支持。

### 3. 全局Token存储

在对接多个第三方`OpenAPI`接口时，通常需要先完成权限校验以获取`Token`。假设需要对接`30`个第三方接口，其中一个接口用于获取`Token`，其余`29`个为业务接口。调用业务接口时，必须在请求头中携带有效的`Token`。

如果每次调用业务接口前都重新获取`Token`，会导致接口调用频繁，同时显著增加代码复杂度。为优化这一流程，常见的做法是将用户名与`Token`绑定后存储到`Redis`中，并设置一个过期时间。这种方法在某些场景下会带来不便，例如，当开发的功能是对接上下游服务的插件，或系统采用强分布式微服务架构时，往往需要将`Redis`打包到镜像中一并部署，从而增加了部署复杂度和维护成本。

在一些对`Token`丢失不敏感的场景下，我们可以将`Token`存储在一个全局变量中。

一个代码使用示例如下所示：

```go
func GetTokenByUserName(userName string) string {
    token, ok := wtoken.Store.RetrieveToken(userName)
    if !ok {
        token = "1a4d0042-4939-433b-9d88-aae75adc37b8"
        wtoken.Store.SaveToken(userName, token, 24*time.Hour)
    }
    return token
}
```

### 4. Gin标准返回结构

在`Gin`接口中，我们可以按照以下方式使用，以下是代码示例：

```go
type User struct {
	Username string `json:"username"`
	Address  string `json:"address"`
}

func main() {
	r := gin.Default()
	r.GET("/user", func(c *gin.Context) {
		user := &User{
			Username: "zhangsan",
			Address:  "caixucun",
		}
		wresp.OK(c, user)
	})
	r.Run()
}
```

调用接口后，返回的结果如下：

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "username": "zhangsan",
        "address": "caixucun"
    }
}
```

对于`message`，应该采用一套标准化的结构，例如定义一组常量进行统一管理，以确保使用时的一致性。

可以在`codes`目录下进行常量的定义，示例如下：

```go
const (
	ParamError        = "参数错误"
	DBConnectionError = "数据库连接错误"
	GatewayError      = "网关错误"
)
```

然后，在代码中可以这样使用`Gin`的工具函数：

```go
wresp.Fail(c, http.StatusBadRequest, codes.ParamError)
```

调用接口后，返回的结果如下：

```json
{
    "code": -1,
    "message": "参数错误"
}
```

这里只展示通过`c.JSON`方法返回`JSON`格式的内容，其他格式如`XML`、`HTML`等不在本示例中讲解。