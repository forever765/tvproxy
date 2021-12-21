package main

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var resp *http.Response
func parseTVB(tvb_type string, isProxy bool, c *gin.Context) (string, error) {
	var myIP string
	var getipErr error
	if isProxy {
		// 写了一个备用的获取ip函数 getIP4TVB，以防第三方api挂掉
		myIP,getipErr = getIP(true, c)
	} else {
		myIP,getipErr = getIP(false, c)
	}
	if getipErr != nil {
		return "",getipErr
	}

	tvbUrl := "https://news.tvb.com/ajax_call/getVideo.php"
	request, err := http.NewRequest("GET", tvbUrl, nil)
	// 特殊方式拼接 Get请求的参数，否则http client不兼容
	q := request.URL.Query()
	q.Add("token", "https://token.tvb.com/stream/live/hls/mobilehd_"+tvb_type+".smil?app=news&feed&client_ip="+myIP)
	request.URL.RawQuery = q.Encode()
	//fmt.Println("TVB获取真实地址的 URL已拼接完成：request.URL.String()")
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

	// Get Real Link
	//var resp *http.Response
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
		fmt.Println("Get TVB Real Link Failed：", string(Body))
		realUrl = ""
	} else {
		result, _ := jsonvalue.UnmarshalString(string(Body))
		n, _ := result.Get("url")
		fmt.Println("Get TVB Real URL Succeed：", n.String())
		// 大陆IP返回的是http协议地址，替换为https。挂了代理就不用https了，因为cdn的tls证书签名有问题???
		//if !isProxy {
			secureLink := strings.Replace(n.String(), "http", "https", 1)
			fmt.Println("TVB Real URL 已替换为 https安全链接：", secureLink)
			realUrl = secureLink
		//} else {
		//	realUrl = n.String()
		//}
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
		getipResp, getipErr = getHTTPClient().Get("http://ip-api.com/json/?lang=zh-CN")
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
	// ip-api要解析json
	if !isProxy {
		result, _ := jsonvalue.UnmarshalString(string(Body))
		n, _ := result.Get("query")
		myIP = n.String()
	} else {
		myIP = string(Body)
	}
	fmt.Println("MyIP (from third party API): ", myIP)
	defer getipResp.Body.Close()
	return myIP,nil
}

// 备用获取本机IP的方式，from tvb live website，也不知道第三方API啥时候会挂掉
func getIP4TVB(isProxy bool, c *gin.Context)(string,error){
	var getipResp *http.Response
	var getipErr error
	if isProxy{
		getipResp, getipErr = getHTTPClientProxy().Get("https://news.tvb.com/live/")
	} else {
		getipResp, getipErr = getHTTPClient().Get("https://news.tvb.com/live/")
	}
	if getipErr != nil {
		return "",getipErr
	}
	Body, err := ioutil.ReadAll(getipResp.Body)
	if err != nil {
		return "",err
	}
	// 从TVB Live官网提取本机IP
	r, _ := regexp.Compile("(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}")
	matchIP := r.FindStringSubmatch(string(Body))
	if err != nil {
		return "",err
	}
	fmt.Println("MyIP (from TVB live website): ", matchIP[0])
	defer getipResp.Body.Close()
	return matchIP[0],nil
}

func tvbHandler(tvb_type string, isProxy bool, c *gin.Context) {
	realM3u8, err := parseTVB(tvb_type, isProxy, c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	if realM3u8 == "" {
		c.AbortWithError(404, errors.New("video not found"))
	} else {
		if isProxy {
			//processedBody := m3u8ProcTVB(realM3u8, baseURL+"i.ts?url=")
			//fmt.Println("processedBody:  ",processedBody)
			//c.Data(200, resp.Header.Get("Content-Type"), []byte(processedBody))

			// 拼接获取 m3u8的link并请求，返回新link，访问新link获取m3u8文件，替换m3u8文件里面的网址
			//c.Redirect(302, "i.ts?url="+realM3u8)
			//https://prd-vcache.edge-global.akamai.tvb.com/__cl/slocalr2526/__c/ott_C_h264/__op/bks/__f/index.m3u8?hdnea=ip=0.0.0.0~st=1640009947~exp=1640096347~acl=/__cl/slocalr2526/__c/ott_C_h264/__op/bks/__f/*~hmac=9a0b292e5f6ffa950a63e8fb1adf675645f35c44712d992b8b439b81ff041c5f
			// 请求自己，获取m3u8索引文件进行处理
			request, err := http.NewRequest("GET", baseURL+"tvb/i.ts?url="+realM3u8, nil)
			resp, err = getHTTPClientProxy().Do(request)
			if err != nil {
				c.AbortWithError(500, err)
			}
			Body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				c.AbortWithError(500, err)
			}

			bodyString := string(Body)
			processedBody := m3u8ProcTVB(bodyString, baseURL+"i.ts?url=", realM3u8)
			//fmt.Println("processedBody:  ",processedBody)
			c.Data(200, resp.Header.Get("Content-Type"), []byte(processedBody))
		} else {
			c.Redirect(302, realM3u8)
		}
	}
}

func m3u8ProcTVB(data string, prefixURL string, linkPerfix string) string {
	var sb strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		l := scanner.Text()
		// 处理不带 # 的行，在前面添加url
		if strings.HasPrefix(l, "#") {
			sb.WriteString(l)
		} else {
			// 避免修改空行
			if l != ""{
				// 从index.m3u8开始全部干掉，保留前段，然后把原内容拼接到最后
				//fmt.Println("url.QueryEscape(l): ", url.QueryEscape(l))
				//fmt.Println("l: ", l)
				reg := regexp.MustCompile("index.*")
				newLink := reg.ReplaceAllString(linkPerfix, l)
				sb.WriteString(baseURL+"i.ts?url="+newLink)
			}

		}
		sb.WriteString("\n")
	}
	//fmt.Println("m3u8ProcTVB 结果 :  ", sb.String())
	return sb.String()
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
	remoteURL := c.FullPath()
	switch remoteURL {
	case "/tvb/inews.m3u8":
		tvbHandler("news_windows1", false, c)
	case "/tvb/inews_proxy.m3u8":
		tvbHandler("news_windows1", true, c)
	case "/tvb/finance.m3u8":
		tvbHandler("financeintl", false, c)
	case "/tvb/finance_proxy.m3u8":
		tvbHandler("financeintl", true, c)
	}
}

func financeHandler(c *gin.Context) {
	remoteURL := c.FullPath()
	switch remoteURL {
		case "/tvb/finance.m3u8":
		tvbHandler("financeintl", false, c)
		case "/tvb/finance_proxy.m3u8":
		tvbHandler("financeintl", true, c)
	}
}
