package main

import (
	"net/http"
	"time"
)

func getHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,  //超时时间
	}
}

func getHTTPClientProxy() *http.Client {
	tr := &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
	}
	return &http.Client{
		Transport: tr,
		Timeout:   time.Second * 10, //超时时间
	}
}
