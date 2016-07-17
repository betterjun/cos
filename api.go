package cos

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/url"
	"os"
	"time"

	"github.com/bitly/go-simplejson"
)

const (
	defaultSignExpireTime = 3600 * 24
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

	url := formatDirectoryURL(c.AppID, bucket, path)
	sign := SignMore(c.AppID, c.SecretID, c.SecretKey, bucket, defaultSignExpireTime)
	return doHttpRequest("POST", url, sign, "application/json", body)
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

	url := formatDirectoryURL(c.AppID, bucket, path)
	url += "?" + params.Encode()
	sign := SignMore(c.AppID, c.SecretID, c.SecretKey, bucket, defaultSignExpireTime)
	return doHttpRequest("GET", url, sign, "application/json", nil)
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

	url := formatDirectoryURL(c.AppID, bucket, path)
	fileid := "/" + c.AppID + "/" + bucket + "/" + path + "/"
	sign := SignOnce(c.AppID, c.SecretID, c.SecretKey, bucket, fileid)
	return doHttpRequest("POST", url, sign, "application/json", body)
}

func (c *COS) QueryFolder(bucket, path string) (err error, jsonResp *simplejson.Json) {
	params := url.Values{}
	params.Add("op", "stat")
	url := formatDirectoryURL(c.AppID, bucket, path)
	url += "?" + params.Encode()
	sign := SignMore(c.AppID, c.SecretID, c.SecretKey, bucket, defaultSignExpireTime)
	return doHttpRequest("GET", url, sign, "application/json", nil)
}

func (c *COS) DeleteFolder(bucket, path string) (err error, jsonResp *simplejson.Json) {
	jsr := simplejson.New()
	jsr.Set("op", "delete")
	body, err := jsr.Encode()
	if err != nil {
		return err, nil
	}

	url := formatDirectoryURL(c.AppID, bucket, path)
	fileid := "/" + c.AppID + "/" + bucket + "/" + path + "/"
	sign := SignOnce(c.AppID, c.SecretID, c.SecretKey, bucket, fileid)
	return doHttpRequest("POST", url, sign, "application/json", body)
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
	writer.Close() // doHttpRequest not defer it, need close it before sending

	url := formatFileURL(c.AppID, bucket, filePath)
	sign := SignMore(c.AppID, c.SecretID, c.SecretKey, bucket, defaultSignExpireTime)
	return doHttpRequest("POST", url, sign, writer.FormDataContentType(), buffer.Bytes())
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

	url := formatFileURL(c.AppID, bucket, path)
	fileid := "/" + c.AppID + "/" + bucket + "/" + path
	sign := SignOnce(c.AppID, c.SecretID, c.SecretKey, bucket, fileid)
	return doHttpRequest("POST", url, sign, "application/json", body)
}

func (c *COS) QueryFile(bucket, path string) (err error, jsonResp *simplejson.Json) {
	params := url.Values{}
	params.Add("op", "stat")

	url := formatFileURL(c.AppID, bucket, path)
	url += "?" + params.Encode()
	sign := SignMore(c.AppID, c.SecretID, c.SecretKey, bucket, defaultSignExpireTime)
	return doHttpRequest("GET", url, sign, "application/json", nil)
}

func (c *COS) DeleteFile(bucket, path string) (err error, jsonResp *simplejson.Json) {
	jsr := simplejson.New()
	jsr.Set("op", "delete")
	body, err := jsr.Encode()
	if err != nil {
		return err, nil
	}

	url := formatFileURL(c.AppID, bucket, path)
	fileid := "/" + c.AppID + "/" + bucket + "/" + path
	sign := SignOnce(c.AppID, c.SecretID, c.SecretKey, bucket, fileid)
	return doHttpRequest("POST", url, sign, "application/json", body)
}

func (c *COS) UploadFileSlice(bucket, filePath, localFileName string) (err error, jsonResp *simplejson.Json) {
	fileHandle, err := os.Open(localFileName)
	if err != nil {
		return err, nil
	}

	hash := sha1.New()
	fileSize, err := io.Copy(hash, fileHandle)
	if err != nil {
		return err, nil
	}

	err, ret := c.createUploadSliceSession(bucket, filePath, fmt.Sprintf("%x", hash.Sum(nil)), fileSize)
	if err != nil || ret.Get("code").MustInt() != 0 {
		return err, nil
	}

	session := ret.Get("data").Get("session").MustString()
	sliceSize := ret.Get("data").Get("slice_size").MustInt64()
	offset := ret.Get("data").Get("offset").MustInt64()
	sliceBuffer := &bytes.Buffer{}
	for {
		_, err = fileHandle.Seek(offset, 0)
		if err != nil {
			return err, nil
		}

		_, err = io.CopyN(sliceBuffer, fileHandle, sliceSize)
		if err != nil && err != io.EOF {
			return err, nil
		}

		// TODO : 根据session和offset开多线程上传
		slice, _ := ioutil.ReadAll(sliceBuffer)
		err, ret = c.uploadSlice(slice, bucket, filePath, session, offset)
		if err != nil || ret.Get("code").MustInt() != 0 {
			return err, nil
		}

		// 以下两种情况为已传完
		if ret.Get("data").Get("access_url").MustString() != "" {
			break
		}

		offset = offset + sliceSize
		if offset >= fileSize {
			break
		}
		sliceBuffer.Reset()
	}

	return nil, ret
}

func (c *COS) createUploadSliceSession(bucket, filePath, sha string, fileSize int64) (err error, jsonResp *simplejson.Json) {
	buffer := &bytes.Buffer{}
	writer := multipart.NewWriter(buffer)
	writer.WriteField("op", "upload_slice")
	writer.WriteField("filesize", fmt.Sprint(fileSize))
	writer.WriteField("sha", sha)
	writer.Close()

	url := formatFileURL(c.AppID, bucket, filePath)
	sign := SignMore(c.AppID, c.SecretID, c.SecretKey, bucket, defaultSignExpireTime)
	return doHttpRequest("POST", url, sign, writer.FormDataContentType(), buffer.Bytes())

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

	url := formatFileURL(c.AppID, bucket, filePath)
	sign := SignMore(c.AppID, c.SecretID, c.SecretKey, bucket, defaultSignExpireTime)
	return doHttpRequest("POST", url, sign, writer.FormDataContentType(), buffer.Bytes())
}
