package cos

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
)

const (
	str             = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	directoryFormat = "http://web.file.myqcloud.com/files/v1/%v/%v/%v/"
	fileFormat      = "http://web.file.myqcloud.com/files/v1/%v/%v/%v"
)

func appSign(appID, secretID, secretKey, bucket, fileid string, expired int64) string {
	// 注意，官方文档上的连接顺序是错的，从其他sdk确认，bucket应放到最后，这文档水平太次了...
	connectedString := fmt.Sprintf("a=%v&k=%v&e=%v&t=%v&r=%v&f=%v&b=%v",
		appID, secretID, expired, time.Now().Unix(), genRandString(32), fileid, bucket)

	data := []byte(connectedString)
	hm := hmac.New(sha1.New, []byte(secretKey))
	hm.Write(data)
	hashValue := hm.Sum(nil)
	hashValue = append(hashValue, data...)

	ret := make([]byte, base64.StdEncoding.EncodedLen(len(hashValue)))
	base64.StdEncoding.Encode(ret, hashValue)
	return string(ret)

	//hashValue := hashfun.HmacSha1([]byte(connectedString), []byte(secretKey))
	//hashValue = append(hashValue, []byte(connectedString)...)
	//return string(b64.Encode(hashValue))
}

func genRandString(size int) string {
	var bytes = make([]byte, size)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = str[v%62]
	}
	return string(bytes)
}

func trimPath(path string) string {
	path = strings.Trim(path, "/")
	if path == "" {
		path = "/"
	}
	return path
}

func getEscapedURL(path string) string {
	u, err := url.Parse(path)
	if err != nil {
		return path
	}
	return u.EscapedPath()
}

func formatDirectoryURL(appID, bucket, path string) string {
	return fmt.Sprintf(directoryFormat, appID, bucket, getEscapedURL(trimPath(path)))
}

func formatFileURL(appID, bucket, path string) string {
	return fmt.Sprintf(fileFormat, appID, bucket, getEscapedURL(trimPath(path)))
}

func doHttpRequest(method, url, sign, contentType string, content []byte) (err error, jsrResp *simplejson.Json) {
	req, err := http.NewRequest(method, url, bytes.NewReader(content))
	if err != nil {
		return fmt.Errorf("create request error: %v", err), nil
	}
	req.Header.Add("Authorization", sign)
	req.Header.Add("Content-Type", contentType)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request error: %v", err), nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response error: %v", err), nil
	}

	jsrResp, err = simplejson.NewJson(body)
	if err != nil {
		return fmt.Errorf("decode response error: %v, Body: %s", err, body), nil
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP StatusCode: %v, Body: %s", resp.StatusCode, body), jsrResp
	}

	return nil, jsrResp
}
