package cos

import (
	"math/rand"
	"testing"
	"time"
)

func TestFunction(t *testing.T) {
	rand.Seed(time.Now().Unix())

	// 在测试开始之前，需要提供如下内容
	AppID := ""
	SecretID := ""
	SecretKey := ""
	Bucket := ""

	if AppID == "" ||
		SecretID == "" ||
		SecretKey == "" ||
		Bucket == "" {
		t.Error("before begin your test, you should define these values: AppID,SecretID,SecretKey,Bucket")
		return
	}

	dirname := "testDir"
	file := dirname + "/" + "testFile.txt"
	localFileName := "D:\\test.txt"

	cosObj := New(AppID, SecretID, SecretKey)
	err, jsr := cosObj.CreateFolder(Bucket, dirname)
	if err == nil {
		t.Log("cosObj.CreateFolder ok")
	} else {
		t.Errorf("cosObj.CreateFolder failed, err=%v", err)
		t.Log(jsr)
	}

	err, jsr = cosObj.UpdateFolder(Bucket, dirname, "test update")
	if err == nil {
		t.Log("cosObj.UpdateFolder ok")
	} else {
		t.Errorf("cosObj.UpdateFolder failed, err=%v", err)
		t.Log(jsr)
	}

	err, jsr = cosObj.QueryFolder(Bucket, dirname)
	if err == nil {
		t.Log("cosObj.QueryFolder ok")
	} else {
		t.Errorf("cosObj.QueryFolder failed, err=%v", err)
		t.Log(jsr)
	}

	err, jsr = cosObj.UploadFile(Bucket, file, localFileName)
	if err == nil {
		t.Log("cosObj.UploadFile ok")
	} else {
		t.Errorf("cosObj.UploadFile failed, err=%v", err)
		t.Log(jsr)
	}

	err, jsr = cosObj.ListFolder(Bucket, dirname, 100, "", 0, "")
	if err == nil {
		t.Log("cosObj.ListFolder ok")
	} else {
		t.Errorf("cosObj.ListFolder failed, err=%v", err)
		t.Log(jsr)
	}

	err, jsr = cosObj.UpdateFile(Bucket, file, "test update file")
	if err == nil {
		t.Log("cosObj.UpdateFile ok")
	} else {
		t.Errorf("cosObj.UpdateFile failed, err=%v", err)
		t.Log(jsr)
	}

	err, jsr = cosObj.QueryFile(Bucket, file)
	if err == nil {
		t.Log("cosObj.QueryFile ok")
	} else {
		t.Errorf("cosObj.QueryFile failed, err=%v", err)
		t.Log(jsr)
	}

	err, jsr = cosObj.DeleteFile(Bucket, file)
	if err == nil {
		t.Log("cosObj.DeleteFile ok")
	} else {
		t.Errorf("cosObj.DeleteFile failed, err=%v", err)
		t.Log(jsr)
	}

	err, jsr = cosObj.UploadFileSlice(Bucket, file, localFileName)
	if err == nil {
		t.Log("cosObj.UploadFileSlice ok")
	} else {
		t.Errorf("cosObj.UploadFileSlice failed, err=%v", err)
		t.Log(jsr)
	}

	err, jsr = cosObj.DeleteFile(Bucket, file)
	if err == nil {
		t.Log("cosObj.DeleteFile ok")
	} else {
		t.Errorf("cosObj.DeleteFile failed, err=%v", err)
		t.Log(jsr)
	}

	err, jsr = cosObj.DeleteFolder(Bucket, dirname)
	if err == nil {
		t.Log("cosObj.DeleteFolder ok")
	} else {
		t.Errorf("cosObj.DeleteFolder failed, err=%v", err)
		t.Log(jsr)
	}
}
