package main

import (
	"net/http"
	"net/url"
	"os"
	"time"
)

func getHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}

func getHTTPClientProxy() *http.Client {
	proxy, _ := url.Parse(os.Getenv("HTTP_PROXY"))
	tr := &http.Transport{
		Proxy:           http.ProxyURL(proxy),
	}
	return &http.Client{
		Transport: tr,
		Timeout:   time.Second * 10, //超时时间
	}
}
