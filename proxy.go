package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"os"
	"time"
)

func tsProxyHandler(c *gin.Context) {
	remoteURL := c.Query("url")

	request, err := http.NewRequest("GET", remoteURL, nil)
	proxy, _ := url.Parse(os.Getenv("HTTP_PROXY"))
	tr := &http.Transport{
		Proxy:           http.ProxyURL(proxy),
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 10, //超时时间
	}
	resp, err := client.Do(request)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	defer resp.Body.Close()
	c.DataFromReader(200, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}
