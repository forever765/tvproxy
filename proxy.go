package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func tsProxyHandler(c *gin.Context) {
	remoteURL := c.Query("url")
	request, err := http.NewRequest("GET", remoteURL, nil)
	resp, err := getHTTPClientProxy().Do(request)
	if err != nil {
		c.AbortWithError(505, err)
		return
	}
	defer resp.Body.Close()
	c.DataFromReader(200, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}
