package main

import (
	"fmt"
	"net/http"
	"time"
)

func getHTTPClient() *http.Client {
	fmt.Println("普通构造：不走代理")
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
