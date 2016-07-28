# cos
qq对象存储服务(COS) golang api.

文档地址
https://www.qcloud.com/doc/product/227

使用示例：
* 建立对象

	cosObj := New(AppID, SecretID, SecretKey)

* 创建文件夹
	err, jsr := cosObj.CreateFolder(Bucket, dirname)
	if err == nil {
		t.Log("cosObj.CreateFolder ok")
	} else {
		t.Errorf("cosObj.CreateFolder failed, err=%v", err)
		t.Log(jsr)
	}

* 更新文件夹属性
	err, jsr = cosObj.UpdateFolder(Bucket, dirname, "test update")
	if err == nil {
		t.Log("cosObj.UpdateFolder ok")
	} else {
		t.Errorf("cosObj.UpdateFolder failed, err=%v", err)
		t.Log(jsr)
	}

* 查询文件夹属性
	err, jsr = cosObj.QueryFolder(Bucket, dirname)
	if err == nil {
		t.Log("cosObj.QueryFolder ok")
	} else {
		t.Errorf("cosObj.QueryFolder failed, err=%v", err)
		t.Log(jsr)
	}

* 上传文件
	err, jsr = cosObj.UploadFile(Bucket, file, localFileName)
	if err == nil {
		t.Log("cosObj.UploadFile ok")
	} else {
		t.Errorf("cosObj.UploadFile failed, err=%v", err)
		t.Log(jsr)
	}

* 列出文件夹下的文件或子目录
	err, jsr = cosObj.ListFolder(Bucket, dirname, 100, "", 0, "")
	if err == nil {
		t.Log("cosObj.ListFolder ok")
	} else {
		t.Errorf("cosObj.ListFolder failed, err=%v", err)
		t.Log(jsr)
	}

* 更新文件属性
	err, jsr = cosObj.UpdateFile(Bucket, file, "test update file")
	if err == nil {
		t.Log("cosObj.UpdateFile ok")
	} else {
		t.Errorf("cosObj.UpdateFile failed, err=%v", err)
		t.Log(jsr)
	}

* 查询文件属性
	err, jsr = cosObj.QueryFile(Bucket, file)
	if err == nil {
		t.Log("cosObj.QueryFile ok")
	} else {
		t.Errorf("cosObj.QueryFile failed, err=%v", err)
		t.Log(jsr)
	}

* 删除文件
	err, jsr = cosObj.DeleteFile(Bucket, file)
	if err == nil {
		t.Log("cosObj.DeleteFile ok")
	} else {
		t.Errorf("cosObj.DeleteFile failed, err=%v", err)
		t.Log(jsr)
	}

* 文件分片上传，适用于较大文件
	err, jsr = cosObj.UploadFileSlice(Bucket, file, localFileName)
	if err == nil {
		t.Log("cosObj.UploadFileSlice ok")
	} else {
		t.Errorf("cosObj.UploadFileSlice failed, err=%v", err)
		t.Log(jsr)
	}

* 删除文件
	err, jsr = cosObj.DeleteFile(Bucket, file)
	if err == nil {
		t.Log("cosObj.DeleteFile ok")
	} else {
		t.Errorf("cosObj.DeleteFile failed, err=%v", err)
		t.Log(jsr)
	}

* 删除目录
	err, jsr = cosObj.DeleteFolder(Bucket, dirname)
	if err == nil {
		t.Log("cosObj.DeleteFolder ok")
	} else {
		t.Errorf("cosObj.DeleteFolder failed, err=%v", err)
		t.Log(jsr)
	}
