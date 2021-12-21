package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func tsProxyHandler(c *gin.Context) {
	remoteURL := c.Query("url")
	fmt.Println("tsProxyHandler收到二次转发请求：",remoteURL)
	request, err := http.NewRequest("GET", remoteURL, nil)
	resp, err := getHTTPClientProxy().Do(request)
	if err != nil {
		c.AbortWithError(505, err)
		return
	}
	defer resp.Body.Close()
	c.DataFromReader(200, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}

func tsProxyHandlerTVB(c *gin.Context) {
	remoteURL := "https://prd-vcache.edge-global.akamai.tvb.com/__cl/slocalr2526/__c/ott_I-NEWS_h264/__op/bks/__f/"+c.FullPath()
	realID := c.Param("id")
	completeRemoteURL := strings.Replace(remoteURL, ":id", realID, -1)
	completeRemoteURL2 := strings.Replace(completeRemoteURL, "/tvb/", "", -1)


	fmt.Println("tsProxyHandler收到二次转发请求222：",completeRemoteURL2)
	request, err := http.NewRequest("GET", completeRemoteURL2, nil)
	resp, err := getHTTPClientProxy().Do(request)
	if err != nil {
		c.AbortWithError(505, err)
		return
	}
	defer resp.Body.Close()
	c.DataFromReader(200, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}