# 短网址服务
使用golang开发的短网址服务

[![github](https://badgen.net/badge/golang/1.12/green)](https://github.com/golang/go)
[![github](https://badgen.net/badge/build/passing/green)](#)
[![github](https://badgen.net/badge/license/GUN/green)](https://github.com/Rejudge-F/ShortLink/blob/master/LICENSE)

Table of Contents
=================

   * [短网址服务](#短网址服务)
   * [Table of Contents](#table-of-contents)
   * [项目目的](#项目目的)
   * [项目架构](#项目架构)
   * [项目设计](#项目设计)
      * [接口设计](#接口设计)
      * [httpServer设计](#httpserver设计)
      * [中间件设计](#中间件设计)
      * [配置设计](#配置设计)
         * [日志配置](#日志配置)
         * [Redis信息配置](#redis信息配置)
   * [如何使用](#如何使用)
   * [短网址服务](#短网址服务-1)
   * [项目目的](#项目目的-1)
   * [项目架构](#项目架构-1)
   * [项目设计](#项目设计-1)
      * [接口设计](#接口设计-1)
      * [httpServer设计](#httpserver设计-1)
      * [中间件设计](#中间件设计-1)
      * [配置设计](#配置设计-1)
         * [日志配置](#日志配置-1)
         * [Redis信息配置](#redis信息配置-1)
   * [如何使用](#如何使用-1)
   * [postman测试结果](#postman测试结果)
      * [POST 测试](#post-测试)
      * [GET 测试](#get-测试)
      * [REDIRECT 测试](#redirect-测试)
      * [LOG 测试](#log-测试)

Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc)

# 项目目的
我们在类似空间微博的地方会遇到字符限制这种问题，但是我们又需要贴一个网址的时候，这时候需要用到短网址服务，短网址的意思就是将一个长网址映射为一个短网址，以此来达到缩短字数的目的

# 项目架构
![结构图](https://github.com/Rejudge-F/ShortLink/blob/master/image/%E6%B5%81%E7%A8%8B.png)

# 项目设计

## 接口设计

主要实现三个接口
```go
// Storage for redis interface
type Storage interface {
	Shorten(url string, exp int64) (string, error)
	ShortlinkInfo(eid string) (interface{}, error)
	UnShorten(eid string) (string, error)
}

```

1. 其中Shorten接口通过redis维护一个自增键来将长网址转换为短网址，具体操作为，先将键自增，然后通过base62来转化成由字母和数字组成的短网址

## httpServer设计
```go
type App struct {
	Router      *mux.Router
	MiddleWares *Middleware
	Config      *Env
}
```
通过mux第三方库来实现路由的功能，然后实现中间件模块，中间件这里用来记录每条请求的执行时间，也可以用来验证等操作，Config为操作Redis的接口

## 中间件设计

```go
type Middleware struct {
}

// LoggingHandler log request with time-out
func (m Middleware) LoggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		timeStart := time.Now()
		next.ServeHTTP(w, r)
		timeEnd := time.Now()
		log.Infof("[%s] %s %v", r.Method, r.URL.String(), timeEnd.Sub(timeStart))
	}
	return http.HandlerFunc(fn)
}

// RecoverHandler recover from panic and return 500
func (m Middleware) RecoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Infof("recover from panic")
				http.Error(w, http.StatusText(500), 500)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
```
通过实现 LoggingHandler来记录每条请求的事件，RecoverHandler来实现服务器panic的时候返回500并记录日志

## 配置设计
### 日志配置
因为记录日志使用的时第三方库seelog，所以配置文件由官方例子改造，放在config文件夹下
```xml
<seelog>
    <outputs formatid="main">
        <filter levels="info,critical,error">
            <console />
        </filter>
        <filter levels="debug, error">
            <file path="./log/App.log" />
        </filter>
    </outputs>
    <formats>
        <format id="main" format="%Date/%Time [%LEV] %Msg%n"/>
    </formats>
</seelog>
```
### Redis信息配置
Redis的操作使用的是第三方库go-redis，配置由env.go进行读取配置，如果没有读到相应的环境变量那么，默认为本地（localhost:6379)的Redis
具体的变量名称为：
- APP_REDIS_ADDR
- APP_REDIS_PASSWD
- APP_REDIS_DB
```go
type Env struct {
	storage Storage
}

func getEnv() *Env {
	addr := os.Getenv("APP_REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	pass := os.Getenv("APP_REDIS_PASSWD")
	if pass == "" {
		pass = ""
	}

	db := os.Getenv("APP_REDIS_DB")
	if db == "" {
		db = "0"
	}
	index, err := strconv.Atoi(db)
	if err != nil {
		log.Debug("%s not db index", db)
		panic(err)
	}

	cli := NewRedisCli(addr, pass, index)

	return &Env{storage: cli}
}
```

# 如何使用
通过三个api来实现
- /api/shorten：通过post一个json
{
	"url":"www.baidu.com",
	"expiration_in_minutes":10
}
来实现生成一个短网址

- /api/info?shortlink={短网址}：用来获取短网址的长网址信息，主要包含
```go
// ShortlinkInfo short_link info
type ShortlinkInfo struct {
	URL                 string `json:"url"`
	CreatedAt           string `json:"created_at"`
	ExpirationInMinutes int64  `json:"expiration_in_minutes"`
}
```
- /{短网址}：用来通过短网址跳转到对应的长网址

# 短网址服务
使用golang开发的短网址服务

[![github](https://badgen.net/badge/golang/1.12/green)](https://github.com/golang/go)
[![github](https://badgen.net/badge/build/passing/green)](#)
[![github](https://badgen.net/badge/license/GUN/green)](https://github.com/Rejudge-F/ShortLink/blob/master/LICENSE)

# 项目目的
我们在类似空间微博的地方会遇到字符限制这种问题，但是我们又需要贴一个网址的时候，这时候需要用到短网址服务，短网址的意思就是将一个长网址映射为一个短网址，以此来达到缩短字数的目的

# 项目架构
![结构图](https://github.com/Rejudge-F/ShortLink/blob/master/image/%E6%B5%81%E7%A8%8B.png)

# 项目设计

## 接口设计

主要实现三个接口
```go
// Storage for redis interface
type Storage interface {
	Shorten(url string, exp int64) (string, error)
	ShortlinkInfo(eid string) (interface{}, error)
	UnShorten(eid string) (string, error)
}

```

1. 其中Shorten接口通过redis维护一个自增键来将长网址转换为短网址，具体操作为，先将键自增，然后通过base62来转化成由字母和数字组成的短网址

## httpServer设计
```go
type App struct {
	Router      *mux.Router
	MiddleWares *Middleware
	Config      *Env
}
```
通过mux第三方库来实现路由的功能，然后实现中间件模块，中间件这里用来记录每条请求的执行时间，也可以用来验证等操作，Config为操作Redis的接口

## 中间件设计

```go
type Middleware struct {
}

// LoggingHandler log request with time-out
func (m Middleware) LoggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		timeStart := time.Now()
		next.ServeHTTP(w, r)
		timeEnd := time.Now()
		log.Infof("[%s] %s %v", r.Method, r.URL.String(), timeEnd.Sub(timeStart))
	}
	return http.HandlerFunc(fn)
}

// RecoverHandler recover from panic and return 500
func (m Middleware) RecoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Infof("recover from panic")
				http.Error(w, http.StatusText(500), 500)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
```
通过实现 LoggingHandler来记录每条请求的事件，RecoverHandler来实现服务器panic的时候返回500并记录日志

## 配置设计
### 日志配置
因为记录日志使用的时第三方库seelog，所以配置文件由官方例子改造，放在config文件夹下
```xml
<seelog>
    <outputs formatid="main">
        <filter levels="info,critical,error">
            <console />
        </filter>
        <filter levels="debug, error">
            <file path="./log/App.log" />
        </filter>
    </outputs>
    <formats>
        <format id="main" format="%Date/%Time [%LEV] %Msg%n"/>
    </formats>
</seelog>
```
### Redis信息配置
Redis的操作使用的是第三方库go-redis，配置由env.go进行读取配置，如果没有读到相应的环境变量那么，默认为本地（localhost:6379)的Redis
具体的变量名称为：
- APP_REDIS_ADDR
- APP_REDIS_PASSWD
- APP_REDIS_DB
```go
type Env struct {
	storage Storage
}

func getEnv() *Env {
	addr := os.Getenv("APP_REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	pass := os.Getenv("APP_REDIS_PASSWD")
	if pass == "" {
		pass = ""
	}

	db := os.Getenv("APP_REDIS_DB")
	if db == "" {
		db = "0"
	}
	index, err := strconv.Atoi(db)
	if err != nil {
		log.Debug("%s not db index", db)
		panic(err)
	}

	cli := NewRedisCli(addr, pass, index)

	return &Env{storage: cli}
}
```

# 如何使用
通过三个api来实现
- /api/shorten：通过post一个json
{
	"url":"www.baidu.com",
	"expiration_in_minutes":10
}
来实现生成一个短网址

- /api/info?shortlink={短网址}：用来获取短网址的长网址信息，主要包含
```go
// ShortlinkInfo short_link info
type ShortlinkInfo struct {
	URL                 string `json:"url"`
	CreatedAt           string `json:"created_at"`
	ExpirationInMinutes int64  `json:"expiration_in_minutes"`
}
```
- /{短网址}：用来通过短网址跳转到对应的长网址

# postman测试结果

## POST 测试
![POST](https://github.com/Rejudge-F/ShortLink/blob/master/image/POST%E6%B5%8B%E8%AF%95%E7%BB%93%E6%9E%9C.png)

## GET 测试
![GET](https://github.com/Rejudge-F/ShortLink/blob/master/image/GET%E6%B5%8B%E8%AF%95%E7%BB%93%E6%9E%9C.png)

## REDIRECT 测试
![Redirect](https://github.com/Rejudge-F/ShortLink/blob/master/image/Redirect.png)

## LOG 测试
![LOG](https://github.com/Rejudge-F/ShortLink/blob/master/image/Log%E5%B1%95%E7%A4%BA.png)


