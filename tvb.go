package main

import (
	"compress/gzip"
	"errors"
	"fmt"
	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var resp *http.Response
func parseTVB(tvb_type string, isProxy bool, c *gin.Context) (string, error) {
	var myIP string
	var getipErr error
	if isProxy {
		myIP,getipErr = getIP(true, c)
	} else {
		myIP,getipErr = getIP(false, c)
	}
	if getipErr != nil {
		return "",getipErr
	}

	tvbUrl := "https://news.tvb.com/ajax_call/getVideo.php"
	request, err := http.NewRequest("GET", tvbUrl, nil)
	q := request.URL.Query()
	q.Add("token", "https://token.tvb.com/stream/live/hls/mobilehd_"+tvb_type+".smil?app=news&feed&client_ip="+myIP)
	request.URL.RawQuery = q.Encode()
	//fmt.Println("TVB获取真实地址的 URL已拼接完成")  // + request.URL.String()
	request.Host = "news.tvb.com"
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("Cache-Control", "no-cache")
	request.Header.Set("Pragma", "no-cache")
	request.Header.Set("sec-ch-ua", " Not A;Brand\";v=\"99\", \"Chromium\";v=\"96\", \"Google Chrome\";v=\"96")
	request.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	request.Header.Set("X-Requested-With", "XMLHttpRequest")
	request.Header.Set("sec-ch-ua-mobile", "?0")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36")
	request.Header.Set("sec-ch-ua-platform", "Windows")
	request.Header.Set("Sec-Fetch-Site", "same-origin")
	request.Header.Set("Sec-Fetch-Mode", "cors")
	request.Header.Set("Sec-Fetch-Dest", "empty")
	request.Header.Set("Referer","https://news.tvb.com/live/")
	request.Header.Set("Accept-Encoding","gzip, deflate, br")
	request.Header.Set("Accept-Language","zh-CN,zh;q=0.9,zh-TW;q=0.8,en;q=0.7")
	request.Header.Set("Cookie","tag_deviceid="+randStr(26)+"; country_code=95882d20a164e8e2c6bb91283bb77bce")

	//var resp *http.Response
	// Get Real Link
	if isProxy {
		resp, err = getHTTPClientProxy().Do(request)
		if err != nil {
			return "",err
		}
	} else {
		resp, _ = getHTTPClient().Do(request)
		if err != nil {
			return "",err
		}
	}
	// unzip data
	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return "", err
	}
	Body, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}
	var realUrl string
	if strings.Contains(string(Body), "error") {
		fmt.Println("TVB真实链接获取失败：", string(Body))
		realUrl = ""
	} else {
		result, _ := jsonvalue.UnmarshalString(string(Body))
		n, _ := result.Get("url")
		fmt.Println("Get TVB Real URL Succeed：", n.String())
		// 大陆IP返回的是http协议地址，替换为https。挂了代理就不用https了，因为cdn的tls证书签名有问题
		if isProxy {
			secureLink := strings.Replace(n.String(), "http", "https", 1)
			fmt.Println("TVB Real URL 已替换为 https安全链接")
			realUrl = secureLink
		} else {
			realUrl = n.String()
		}
	}

	defer resp.Body.Close()
	return realUrl, nil
}

func getIP(isProxy bool, c *gin.Context)(string,error){
	var getipResp *http.Response
	var getipErr error
	var myIP string
	if isProxy{
		getipResp, getipErr = getHTTPClientProxy().Get("https://api.ipify.org")
	} else {
		fmt.Println("不走代理")
		//request2, _ := http.NewRequest("GET", "http://ip-api.com/json/?lang=zh-CN", nil)
		//client := &http.Client{
		//	Timeout:   time.Second * 10, //超时时间
		//}
		//getipResp, getipErr = client.Do(request2)
		getipResp, getipErr = getHTTPClient().Get("http://ip-api.com/json/?lang=zh-CN")
		//getipResp, getipErr = getHTTPClient().Get("https://www.google.com")
	}
	if getipErr != nil {
		return "",getipErr
	}
	Body, err := ioutil.ReadAll(getipResp.Body)
	if err != nil {
		return "",err
	}
	if string(Body) == "" {
		c.AbortWithError(404, errors.New("get ip failed, pleace check api"))
	}
	fmt.Println(string(Body))
	// ip-api要解析json
	if !isProxy {
		result, _ := jsonvalue.UnmarshalString(string(Body))
		n, _ := result.Get("query")
		myIP = n.String()
	} else {
		myIP = string(Body)
	}
	fmt.Println("MyIP: ", myIP)
	defer getipResp.Body.Close()
	return myIP,nil
}

func tvbHandler(liveName string, c *gin.Context) {
	var tvb_type string
	var isProxy bool = false
	switch liveName {
	case "inews":
		tvb_type = "news_windows1"
	case "inews_proxy":
		tvb_type = "news_windows1"
		isProxy = true
	case "finance":
		tvb_type = "financeintl"
	case "finance_proxy":
		tvb_type = "financeintl"
		isProxy = true
	default:
		tvb_type = "news_windows1"
	}
	realM3u8, err := parseTVB(tvb_type, isProxy, c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	if realM3u8 == "" {
		c.AbortWithError(404, errors.New("video not found"))
	} else {
		if isProxy {
			processedBody := m3u8Proc(realM3u8, baseURL+"i.ts?url=")
			c.Data(200, resp.Header.Get("Content-Type"), []byte(processedBody))
			//tsProxyHandlerThin(realM3u8, c)
		} else {
			c.Redirect(302, realM3u8)
		}
	}
}

func randStr(length int) string {
	str := "0123456789"
	//str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	rand.Seed(time.Now().UnixNano()+ int64(rand.Intn(100)))
	for i := 0; i < length; i++ {
		result = append(result, bytes[rand.Intn(len(bytes))])
	}
	return string(result)
}

func iNewsHandler(c *gin.Context) {
	tvbHandler("inews", c)
}

func financeHandler(c *gin.Context) {
	tvbHandler("j5_ch85", c)
}
