package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	//_ "github.com/joho/godotenv/autoload"
)

var baseURL string
var proxyURL *url.URL
var err error

func main() {
	// init
	fmt.Println("TVProxy Started (https://github.com/forever765/tvproxy)")
	listenOn := os.Getenv("TVPROXY_LISTEN")
	if listenOn == "" {
		listenOn = "127.0.0.1:10086"
	}
	baseURL = os.Getenv("TVPROXY_BASE_URL")
	if baseURL == "" {
		baseURL = "http://" + listenOn + "/"
	}
	proxyURL, err = url.Parse(os.Getenv("TVPROXY_HTTP_PROXY"))
	if err != nil {
		fmt.Println("proxy解析失败，请尝试添加http://")
	}
	// webserver
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/i.ts", tsProxyHandler)
	r.GET("/iptv.m3u", m3uHandler)
	tvb := r.Group("/tvb")
	{
		tvb.GET("/inews.m3u8", iNewsHandler)
		tvb.GET("/inews_proxy.m3u8", iNewsHandler)
		tvb.GET("/finance.m3u8", financeHandler)
		tvb.GET("/finance_proxy.m3u8", financeHandler)
		tvb.GET("/2/:id/index.m3u8", tsProxyHandler)
	}
	rthk := r.Group("/rthk")
	{
		rthk.GET("/31.m3u8", rthk31Handler)
		rthk.GET("/32.m3u8", rthk32Handler)
	}
	r.Run(listenOn)
}
