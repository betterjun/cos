package cos

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
)

const (
	COS_HOST            = "http://web.file.myqcloud.com"
	Default_Expire_Time = 3600 * 24 * 30
)

// 生成多次有效的sign，签名中不绑定文件id，需要设置大于当前时间的有效期，有效期内此签名可多次使用，有效期最长可设置三个月。
func SignMore(appID, secretID, secretKey, bucket string, duration int64) string {
	expired := time.Now().Unix() + duration
	return appSign(appID, secretID, secretKey, bucket, "", expired)
}

// 生成单次有效的sign，签名中绑定文件id，有效期必须设置为0，此签名只可使用一次，且只能应用于被绑定的文件。
func SignOnce(appID, secretID, secretKey, bucket, fileid string) string {
	return appSign(appID, secretID, secretKey, bucket, fileid, 0)
}

type COS struct {
	AppID     string
	SecretID  string
	SecretKey string
}

func New(appID, secretID, secretKey string) *COS {
	return &COS{AppID: appID, SecretID: secretID, SecretKey: secretKey}
}

func (c *COS) CreateFolder(bucket, path string) (err error, jsonResp *simplejson.Json) {
	jsr := simplejson.New()
	jsr.Set("op", "create")
	body, err := jsr.Encode()
	if err != nil {
		return err, nil
	}

	url := formatURL(c.AppID, bucket, path)
	fmt.Println("url is ", url)
	sign := SignMore(c.AppID, c.SecretID, c.SecretKey, bucket, Default_Expire_Time)
	return do("POST", url, sign, "application/json", body)
}

/*
path 暂不支持前缀查询
num	是	Int	拉取的总数
pattern	否	String	eListBoth,eListDirOnly,eListFileOnly默认both
order	否	Int	默认正序(=0), 填1为反序,
context	否	String	透传字段，查看第一页，则传空字符串。若需要翻页，需要将前一页返回值中的context透传到参数中。order用于指定翻页顺序。若order填0，则从当前页正序/往下翻页；若order填1，则从当前页倒序/往上翻页
*/
func (c *COS) ListFolder(bucket, path string, num uint64, pattern string, order int8, context string) (err error, jsonResp *simplejson.Json) {
	params := url.Values{}
	params.Add("op", "list")

	if num <= 0 {
		num = 20
	}
	params.Add("num", fmt.Sprint(num))
	// 以下参数，若无效则忽略，采用服务器默认值
	if pattern == "eListBoth" || pattern == "eListDirOnly" || pattern == "eListFileOnly" {
		params.Add("pattern", pattern)
	}
	if order == 0 || order == 1 {
		params.Add("order", fmt.Sprint(order))
	}
	if context != "" {
		params.Add("context", context)
	}

	url := formatURL(c.AppID, bucket, path)
	url += "?" + params.Encode()
	fmt.Println("url is ", url)
	sign := SignMore(c.AppID, c.SecretID, c.SecretKey, bucket, Default_Expire_Time)
	return do("GET", url, sign, "application/json", nil)
}

func (c *COS) UpdateFolder(bucket, path, attribute string) (err error, jsonResp *simplejson.Json) {
	jsr := simplejson.New()
	jsr.Set("op", "update")
	if attribute != "" {
		jsr.Set("biz_attr", attribute)
	}

	body, err := jsr.Encode()
	if err != nil {
		return err, nil
	}

	url := formatURL(c.AppID, bucket, path)
	fmt.Println("url is ", url)
	fileid := "/" + c.AppID + "/" + bucket + "/" + path + "/"
	sign := SignOnce(c.AppID, c.SecretID, c.SecretKey, bucket, fileid)
	return do("POST", url, sign, "application/json", body)
}

func (c *COS) QueryFolder(bucket, path string) (err error, jsonResp *simplejson.Json) {
	params := url.Values{}
	params.Add("op", "stat")

	url := formatURL(c.AppID, bucket, path)
	url += "?" + params.Encode()
	fmt.Println("url is ", url)
	sign := SignMore(c.AppID, c.SecretID, c.SecretKey, bucket, Default_Expire_Time)
	return do("GET", url, sign, "application/json", nil)
}

func (c *COS) DeleteFolder(bucket, path string) (err error, jsonResp *simplejson.Json) {
	jsr := simplejson.New()
	jsr.Set("op", "delete")

	body, err := jsr.Encode()
	if err != nil {
		return err, nil
	}

	url := formatURL(c.AppID, bucket, path)
	fmt.Println("url is ", url)
	fileid := "/" + c.AppID + "/" + bucket + "/" + path + "/"
	sign := SignOnce(c.AppID, c.SecretID, c.SecretKey, bucket, fileid)
	return do("POST", url, sign, "application/json", body)
}

func (c *COS) UploadFile(bucket, filePath, localFileName string) (err error, jsonResp *simplejson.Json) {
	fileHandle, err := os.Open(localFileName)
	if err != nil {
		return err, nil
	}
	fileContent, err := ioutil.ReadAll(fileHandle)
	if err != nil {
		return err, nil
	}

	buffer := &bytes.Buffer{}
	writer := multipart.NewWriter(buffer)
	writer.WriteField("op", "upload")
	sha := fmt.Sprintf("%x", sha1.Sum(fileContent))
	writer.WriteField("sha", sha)

	fcField, _ := writer.CreateFormField("filecontent")
	_, err = fcField.Write(fileContent)
	if err != nil {
		return err, nil
	}
	writer.Close() // do not defer it, need close it before sending

	url := formatURL(c.AppID, bucket, filePath)
	url = strings.TrimSuffix(url, "/")
	fmt.Println("url is ", url)
	sign := SignMore(c.AppID, c.SecretID, c.SecretKey, bucket, Default_Expire_Time)
	return do("POST", url, sign, writer.FormDataContentType(), buffer.Bytes())
}

func (c *COS) UpdateFile(bucket, path, attribute string) (err error, jsonResp *simplejson.Json) {
	jsr := simplejson.New()
	jsr.Set("op", "update")
	if attribute != "" {
		jsr.Set("biz_attr", attribute)
	}

	body, err := jsr.Encode()
	if err != nil {
		return err, nil
	}

	url := formatURL(c.AppID, bucket, path)
	url = strings.TrimSuffix(url, "/")
	fmt.Println("url is ", url)
	fileid := "/" + c.AppID + "/" + bucket + "/" + path
	sign := SignOnce(c.AppID, c.SecretID, c.SecretKey, bucket, fileid)
	return do("POST", url, sign, "application/json", body)
}

func (c *COS) QueryFile(bucket, path string) (err error, jsonResp *simplejson.Json) {
	params := url.Values{}
	params.Add("op", "stat")

	url := formatURL(c.AppID, bucket, path)
	url = strings.TrimSuffix(url, "/")
	url += "?" + params.Encode()
	fmt.Println("url is ", url)
	sign := SignMore(c.AppID, c.SecretID, c.SecretKey, bucket, Default_Expire_Time)
	return do("GET", url, sign, "application/json", nil)
}

func (c *COS) DeleteFile(bucket, path string) (err error, jsonResp *simplejson.Json) {
	jsr := simplejson.New()
	jsr.Set("op", "delete")

	body, err := jsr.Encode()
	if err != nil {
		return err, nil
	}

	url := formatURL(c.AppID, bucket, path)
	url = strings.TrimSuffix(url, "/")
	fmt.Println("url is ", url)
	fileid := "/" + c.AppID + "/" + bucket + "/" + path
	sign := SignOnce(c.AppID, c.SecretID, c.SecretKey, bucket, fileid)
	return do("POST", url, sign, "application/json", body)
}

func (c *COS) UploadFileSlice(bucket, filePath, localFileName string) (err error, jsonResp *simplejson.Json) {
	fileHandle, err := os.Open(localFileName)
	if err != nil {
		return err, nil
	}
	fileContent, err := ioutil.ReadAll(fileHandle)
	if err != nil {
		return err, nil
	}

	sha := fmt.Sprintf("%x", sha1.Sum(nil))
	fileSize := int64(len(fileContent))
	err, ret := c.createUploadSliceSession(bucket, filePath, sha, fileSize)

	var session string
	var offset int64
	var sliceSize int64
	for {
		if err != nil || ret.Get("code").MustInt() != 0 {
			return err, nil
		}

		retData := ret.Get("data")
		if retData.Get("url").MustString() != "" { // 已传完
			break
		}

		if session == "" {
			session = retData.Get("session").MustString()
		}
		if offset == 0 {
			offset = retData.Get("offset").MustInt64()
		}
		if sliceSize == 0 {
			retData.Get("slice_size").MustInt64()
		}

		err, ret = c.uploadSlice(fileContent[offset:offset+sliceSize+1], bucket, filePath, session, offset)
		offset = offset + sliceSize
		if offset >= fileSize {
			break
		}
	}
	return nil, nil
}

func (c *COS) createUploadSliceSession(bucket, filePath, sha string, fileSize int64) (err error, jsonResp *simplejson.Json) {
	buffer := &bytes.Buffer{}
	writer := multipart.NewWriter(buffer)
	writer.WriteField("op", "upload_slice")
	writer.WriteField("filesize", fmt.Sprint(fileSize))
	writer.WriteField("sha", sha)
	writer.Close()

	url := formatURL(c.AppID, bucket, filePath)
	url = strings.TrimSuffix(url, "/")
	fmt.Println("url is ", url)
	sign := SignMore(c.AppID, c.SecretID, c.SecretKey, bucket, Default_Expire_Time)
	return do("POST", url, sign, writer.FormDataContentType(), buffer.Bytes())
}

func (c *COS) uploadSlice(fileContent []byte, bucket, filePath, session string, offset int64) (err error, jsonResp *simplejson.Json) {
	buffer := &bytes.Buffer{}
	writer := multipart.NewWriter(buffer)
	writer.WriteField("op", "upload_slice")
	writer.WriteField("sha", fmt.Sprintf("%x", sha1.Sum(fileContent)))
	writer.WriteField("session", session)
	writer.WriteField("offset", fmt.Sprint(offset))
	fcField, _ := writer.CreateFormField("filecontent")
	_, err = fcField.Write(fileContent)
	if err != nil {
		return err, nil
	}
	writer.Close()

	url := formatURL(c.AppID, bucket, filePath)
	url = strings.TrimSuffix(url, "/")
	fmt.Println("url is ", url)
	sign := SignMore(c.AppID, c.SecretID, c.SecretKey, bucket, Default_Expire_Time)
	return do("POST", url, sign, writer.FormDataContentType(), buffer.Bytes())
}
