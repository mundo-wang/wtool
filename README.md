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

### 2. `HTTP`工具

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

### 3. 全局`Token`存储

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

### 4. `Gin`标准返回结构

文件中按照以下方式新建错误码：

```go
var (
    InvalidInput = wresp.NewErrorCode(10003, "提交的数据格式无效，请检查输入的内容")
    Unauthorized = wresp.NewErrorCode(10004, "未登录或权限不足，无法访问此资源")
    Forbidden    = wresp.NewErrorCode(10005, "访问被拒绝，您没有权限操作此资源，请联系管理员")
    NotFound     = wresp.NewErrorCode(10006, "请求的资源未找到，请确认URL是否正确")
    Timeout      = wresp.NewErrorCode(10007, "请求超时，请稍后重试")
)
```

错误码是面向前端展示给用户的关键信息。由于用户通常缺乏技术背景，他们依赖错误信息来理解问题发生的原因。因此，为了提升用户体验，错误码应具有足够的区分度，以便用户能够查阅相关文档或向后台人员反馈，从而更高效地定位和解决问题。错误信息应简洁明了，避免使用过于技术化的术语，而要清晰地传达问题的本质原因。

为确保错误码的规范化管理，建议使用纯数字并按业务模块进行分组。这种分组方式有助于简化错误码的管理和查找，显著提高问题定位和排查效率。通过这种设计，错误码系统能够更好地支持业务需求，并与用户高效沟通。

我们先创建以下两个错误码：

```go
UserNotFound     = wresp.NewErrorCode(10008, "未找到对应用户，请检查用户是否存在")
CreateUserFailed = wresp.NewErrorCode(10009, "创建用户时出错，请检查创建参数")
```

在这里，错误码应该尽量细化，为每一种错误类型分配一个独立的错误码，同时编写清晰、易于理解的错误信息。

错误码既可应用于`service`层代码，也可应用于`controller`层代码。我们为`service`层的两个方法添加简洁的错误判断与返回：

```go
type User struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type UserService struct {
}

func (u *UserService) GetUserInfo(id int64) (*User, error) {
	if id == 10 {
		return nil, wresp.UserNotFound
	}
	user := &User{
		Id:   id,
		Name: "zhangsan",
	}
	return user, nil
}

func (u *UserService) CreateUser(user *User) error {
	return wresp.CreateUserFailed
}
```

接下来对`controller`层的代码进行修改，具体改动如下：

```go
type UserApi struct {
	service.UserService
}

func GetUserApi() *UserApi {
	return &UserApi{}
}

func (u *UserApi) GetUserInfo(c *gin.Context) (interface{}, error) {
	user, err := u.UserService.GetUserInfo(10)
	if err != nil {
        if !wresp.IsErrorCode(err) {
			wlog.Error("call u.UserService.GetUserInfo failed").Err(err).Field("req", req).Log()
		}
		return nil, err
	}
	return user, nil
}

func (u *UserApi) CreateUser(c *gin.Context) (interface{}, error) {
	user := &service.User{
		Id:   20,
		Name: "lisi",
	}
	err := u.UserService.CreateUser(user)
	if err != nil {
		if !wresp.IsErrorCode(err) {
			wlog.Error("call u.UserService.CreateUser failed").Err(err).Field("req", req).Log()
		}
		return nil, err
	}
	return nil, nil
}
```

可以看到，我们将两个`Gin`接口函数改造为包装后的方法，这样`controller`层可以直接返回`service`层返回的具体错误码对象（透传），并交由“`Gin`进阶返回结构”工具进行处理与返回。

在`controller`层直接打印`service`层返回的错误码并不合理。此类错误通常源于用户的不当操作，若遭遇恶意攻击，可能会导致系统生成大量`ERROR`级别的日志，干扰正常监控。因此，可以使用`IsErrorCode`函数进行判断：对于业务错误，不记录日志；对于系统错误，则记录日志，以确保系统错误的可追溯性。当前方案仍存在一定不便，后续若有更优解，再进行优化。

对于`router`部分的代码逻辑，改动如下所示：

```go
func SetRouter(s *wresp.Server) {
    r := s.Router
	users := r.Group("/api/v1/users")
	{
		users.GET("/get", s.WrapHandler(api.GetUsersApi().GetUserInfo))
		users.POST("/set", s.WrapHandler(api.GetUsersApi().CreateUser))
	}
    s.Router = r
}
```

在每个`router`部分函数的开头，需要从`*wresp.Server`类型的对象`s`中获取`router`对象。完成路由注册后，再将`router`送回`s`。此外，应使用`s.WrapHandler`方法封装`controller`层的方法，以便其返回结果能够直接由工具处理。

接下来是主函数部分的修改，我们把`router`对象的创建逻辑从`router`目录移交到了主函数文件的`NewServer`函数中：

```go
func main() {
	s := NewServer()
	err := s.Router.Run(":8081")
	if err != nil {
		wlog.Error("call r.Run failed").Err(err).Field("port", 8081).Log()
		return
	}
}

func NewServer() *wresp.Server {
	s := &wresp.Server{
		Router: gin.Default(),
	}
	router.SetRouter(s)
	return s
}
```

如果存在多个路由注册函数，可以在`NewServer`函数中统一调用它们进行注册：

```go
func NewServer() *wresp.Server {
	s := &wresp.Server{
		Router: gin.Default(),
	}
	router.SetRouter1(s)
	router.SetRouter2(s)
	router.SetRouter3(s)
	return s
}
```

对于中间件的编写，我们使用到了`wresp.MiddlewareWrapper`这个函数类型，具体代码如下：

```go
func MiddlewareA() wresp.MiddlewareWrapper {
	return func(c *gin.Context) error {
		fmt.Println("MiddlewareA - Before Next")
		if c.Query("userName") == "admin" {
			return code.UserNameAlreadyExist
		}
		c.Next()
		fmt.Println("MiddlewareA - After Next")
		return nil
	}
}

func MiddlewareB() wresp.MiddlewareWrapper {
	return func(c *gin.Context) error {
		fmt.Println("MiddlewareB - Before Next")
		c.Next()
		fmt.Println("MiddlewareB - After Next")
		return nil
	}
}
```

值得注意的是，使用上述代码后，我们无需手动调用`c.Abort()`，只需返回错误即可。

注册中间件时，使用到了`WrapMiddleware`方法，代码如下：

```go
r.Use(s.WrapMiddleware(MiddlewareA()))
r.Use(s.WrapMiddleware(MiddlewareB()))
```

这样改造后，中间件代码也能返回错误码格式的`error`了。

### 5. 联系方式

如有任何问题或建议，请通过以下方式联系我：

> - 邮箱：userwsj@126.com
> - 微信：13136163259

感谢您的积极反馈！