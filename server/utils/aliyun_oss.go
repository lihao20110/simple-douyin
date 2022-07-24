package utils

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/lihao20110/simple-douyin/server/global"
	"go.uber.org/zap"
)

func NewBucket() (*oss.Bucket, error) {
	//  Endpoint以杭州为例，其它Region请按实际情况填写。
	//  endpoint := "http://oss-cn-hangzhou.aliyuncs.com"
	//  阿里云账号AccessKey拥有所有API的访问权限，风险很高。强烈建议您创建并使用RAM用户进行API访问或日常运维，请登录RAM控制台创建RAM用户。
	//    accessKeyId := "<yourAccessKeyId>"
	//    accessKeySecret := "<yourAccessKeySecret>"

	// 创建OSSClient实例。
	client, err := oss.New(global.DouYinCONFIG.AliyunOSS.Endpoint, global.DouYinCONFIG.AliyunOSS.AccessKeyId, global.DouYinCONFIG.AliyunOSS.AccessKeySecret)
	if err != nil {
		return nil, err
	}

	// 获取存储空间。
	bucket, err := client.Bucket(global.DouYinCONFIG.AliyunOSS.BucketName)
	if err != nil {
		return nil, err
	}

	return bucket, nil
}

//MultipartUpload 分片上传
func MultipartUpload(objectName, localFileName string) (string, error) {
	bucket, err := NewBucket()
	if err != nil {
		global.DouYinLOG.Error("function AliyunOSS.NewBucket() Failed", zap.Any("err", err.Error()))
		return "", errors.New("function AliyunOSS.NewBucket() Failed, err:" + err.Error())
	}
	//本地文件分片
	chunks, err := oss.SplitFileByPartNum(localFileName, global.DouYinCONFIG.AliyunOSS.PartNum)
	fd, err := os.Open(localFileName)
	defer fd.Close()

	// 指定过期时间。
	expires := time.Date(2049, time.January, 10, 23, 0, 0, 0, time.UTC)

	// 如果需要在初始化分片时设置请求头，请参考以下示例代码。
	options := []oss.Option{
		oss.MetadataDirective(oss.MetaReplace),
		oss.Expires(expires),
		// 指定该Object被下载时的网页缓存行为。
		// oss.CacheControl("no-cache"),
		// 指定该Object被下载时的名称。
		// oss.ContentDisposition("attachment;filename=FileName.txt"),
		// 指定该Object的内容编码格式。
		// oss.ContentEncoding("gzip"),
		// 指定对返回的Key进行编码，目前支持URL编码。
		// oss.EncodingType("url"),
		// 指定Object的存储类型。
		// oss.ObjectStorageClass(oss.StorageStandard),
	}
	// 步骤1：初始化一个分片上传事件，并指定存储类型为标准存储。
	imur, err := bucket.InitiateMultipartUpload(objectName, options...)
	// 步骤2：上传分片。
	var parts []oss.UploadPart
	for _, chunk := range chunks {
		fd.Seek(chunk.Offset, 0)
		// 调用UploadPart方法上传每个分片。
		part, err := bucket.UploadPart(imur, fd, chunk.Size, chunk.Number)
		if err != nil {
			global.DouYinLOG.Error("UploadPart failed", zap.Error(err))
			return "", err
		}
		parts = append(parts, part)
	}
	// 指定Object的读写权限为公共读，默认为继承Bucket的读写权限。
	objectAcl := oss.ObjectACL(oss.ACLPublicRead)

	// 步骤3：完成分片上传，指定文件读写权限为公共读。
	cmur, err := bucket.CompleteMultipartUpload(imur, parts, objectAcl)
	if err != nil {
		global.DouYinLOG.Error("UploadPart failed", zap.Error(err))
		return "", err
	}
	fmt.Println("cmur:", cmur)
	return global.DouYinCONFIG.AliyunOSS.BucketUrl + "/" + objectName, nil
}

//UploadFromFile 上传文件
func UploadFromFile(objectName, localFileName string) (string, error) {
	bucket, err := NewBucket()
	if err != nil {
		global.DouYinLOG.Error("function AliyunOSS.NewBucket() Failed", zap.Any("err", err.Error()))
		return "", errors.New("function AliyunOSS.NewBucket() Failed, err:" + err.Error())
	}
	err = bucket.PutObjectFromFile(objectName, localFileName)
	if err != nil {
		global.DouYinLOG.Error("PutObjectFromFile Failed", zap.Any("err", err.Error()))
		return "", errors.New("PutObjectFromFile Failed, err:" + err.Error())
	}
	return global.DouYinCONFIG.AliyunOSS.BucketUrl + "/" + objectName, nil
}
