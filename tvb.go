package main

import (
	"compress/gzip"
	"errors"
	"fmt"
	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func parseTVB(liveName string, c *gin.Context) (string, error) {
	fmt.Println("开始执行tvb代码块")
	var tvb_type string
	if liveName == "inews" {
		tvb_type = "news_windows1"
	} else {
		tvb_type = "financeintl"
	}
	myIP,err := getIP(c)
	if err != nil {
		return "",err
	}

	client := http.Client{} //getHTTPClient()
	tvbUrl := "https://news.tvb.com/ajax_call/getVideo.php"
	request, err := http.NewRequest("GET", tvbUrl, nil)
	q := request.URL.Query()
	q.Add("token", "http://token.tvb.com/stream/live/hls/mobilehd_"+tvb_type+".smil?app=news&feed&client_ip="+myIP)
	request.URL.RawQuery = q.Encode()
	fmt.Println("TVB获取真实地址的 URL已拼接完成")  //request.URL.String()

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
	request.Header.Set("Cookie","tag_deviceid=28987705633044241626698564; country_code=95882d20a164e8e2c6bb91283bb77bce")
	resp, _ := client.Do(request)
	if err != nil {
		return "",err
	}
	//proxy, _ := url.Parse("http://192.168.123.66:7890")
	//tr := &http.Transport{
	//	Proxy:           http.ProxyURL(proxy),
	//}
	//
	//client := &http.Client{
	//	Transport: tr,
	//	Timeout:   time.Second * 10, //超时时间
	//}
	//resp, err := client.Do(request)
	//if err != nil {
	//	return "",err
	//}
	//fmt.Println(request.Header)
	reader, _ := gzip.NewReader(resp.Body)
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
		fmt.Println("真实 URL获取成功：", n.String())
		realUrl = n.String()
	}
	defer resp.Body.Close()
	return realUrl, nil
}

func getIP(c *gin.Context)(string,error){
	request, err := http.NewRequest("GET", "https://api.ipify.org", nil)
	//getip_resp, err := getHTTPClient().Get("http://ip.3322.net/")
	if err != nil {
		return "",err
	}

	//proxy, _ := url.Parse("http://192.168.123.66:7890")
	//tr := &http.Transport{
	//	Proxy:           http.ProxyURL(proxy),
	//}

	client := &http.Client{
		//Transport: tr,
		Timeout:   time.Second * 10, //超时时间
	}
	getipResp, err := client.Do(request)
	if err != nil {
		return "",err
	}
	Body, err := ioutil.ReadAll(getipResp.Body)
	if err != nil {
		return "",err
	}
	//defer getipResp.Body.Close()
	if string(Body) == "" {
		c.AbortWithError(404, errors.New("get ip failed"))
	}
	pureIP := strings.Replace(string(Body), "\n", "", -1)
	fmt.Println("MyIP: ", pureIP)
	return pureIP,nil
}

func tvbHandler(liveName string, c *gin.Context) {
	readM3u8, err := parseTVB(liveName, c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	if readM3u8 == "" {
		c.AbortWithError(404, errors.New("video not found"))
	} else {
		c.Redirect(302, readM3u8)
	}
}

func iNewsHandler(c *gin.Context) {
	tvbHandler("inews", c)
}

func financeHandler(c *gin.Context) {
	tvbHandler("j5_ch85", c)
}
