package main

import (
	"encoding/json"
	core "exec/UPLStudy12_15/Study12_19/config"
	"fmt"
	"github.com/spf13/viper"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync/atomic"
	"time"
)

// var configFiles = []string{"config.yaml"}
var config *viper.Viper
var times *int64
var timeNew int64

//var configCacheString = map[string]string{}
//var configCacheInt = map[string]int{}
//var configCacheDuration = map[string]time.Duration{}
//var configCacheFloat64 = map[string]float64{}

func init() {
	config = viper.New()
	config.SetConfigType("yaml")
	//config.AddConfigPath("./config/")
	config.SetConfigFile("D:\\go开发\\exec\\apps\\UPLStudy12_15\\Study12_19\\config.yaml")
	var err error
	err = config.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}
	var timeOld = time.Now().UnixMilli()
	times = &timeOld
}

var tokenBucket chan struct{}

func fillToken() {
	// 加载配置文件
	capacity := config.GetInt("capacity")
	fmt.Println(capacity)
	fillInterval := time.Second * (config.GetDuration("fillInterval"))
	//var fillInterval = time.Second * 2
	//var capacity = 100
	tokenBucket = make(chan struct{}, capacity)
	ticker := time.NewTicker(fillInterval)
	for {
		select {
		case <-ticker.C:
			select {
			case tokenBucket <- struct{}{}:
			default:
			}
			//fmt.Println("current token cnt:", len(tokenBucket), time.Now())
		}
	}
}
func echo(wr http.ResponseWriter, r *http.Request) {
	msg, err := io.ReadAll(r.Body)
	if err != nil {
		wr.Write([]byte("echo error"))
		return
	}

	writeLen, err := wr.Write(msg)
	if err != nil || writeLen != len(msg) {
		log.Println(err, "write len:", writeLen)
	}
	data, _ := json.Marshal(Result{
		Code: 200,
		Msg:  "成功",
		Bug:  "",
	})
	wr.Write(data)
}

func html404(wr http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("404.html")
	if err != nil {
		log.Println("err:", err)
		return
	}
	t.Execute(wr, nil)
}

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Bug  string `json:"bug"`
}

// 处理的方法
// 使用适配器模式编写一个接口限流器
func interfaceLimitMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		//时间戳
		//比较时间
		//进来表示能够被执行，但是不一定有令牌，所以我需要先看有没有令牌
		// 记录的原子当前时间-请求到来的时间<（当前记录的时间-令牌创建的时间） 时 快速失败
		//让用户选择策略。第一种就是快速失败，第二种就是默认的等待阻塞
		// 1、快速失败
		timeNew = time.Now().UnixMilli()
		if len(tokenBucket) <= 0 {
			if (time.Now().UnixMilli() - *times) < config.GetInt64("fillInterval")*1000 {
				atomic.StoreInt64(times, timeNew)
				// 这里两种情况 https://mojotv.cn/2019/07/30/golang-http-request
				u := &url.URL{
					Scheme: "http",
					Host:   "localhost:8080",
				}
				proxy := httputil.NewSingleHostReverseProxy(u)
				request.URL.Path = "/404.html"
				proxy.ServeHTTP(writer, request)
				return
				// 2、阻塞等待
			}
		}
		atomic.StoreInt64(times, timeNew)
		<-tokenBucket
		handler.ServeHTTP(writer, request)
	})
}

// 这里把限流中间件和日志中间件分开
func logLimitMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		filePath := "D:\\go开发\\exec\\apps\\log.log"
		file, _ := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE, 0666)
		file.Write([]byte("路径 " + request.URL.Path + "  "))
		start := time.Now()
		handler.ServeHTTP(writer, request)
		file.Write([]byte(fmt.Sprintf("%v", time.Since(start)) + "\n"))
	})
}

func main() {
	go fillToken()
	////有请求进来就取令牌
	r := core.NewRouter()
	r.Use(interfaceLimitMiddleware)
	r.Use(logLimitMiddleware)
	http.Handle("/", r.Chain(http.HandlerFunc(echo)))
	http.Handle("/404.html", r.Chain(http.HandlerFunc(html404)))
	http.ListenAndServe("localhost:8080", nil)
}
