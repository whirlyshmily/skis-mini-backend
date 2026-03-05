package services

import (
	"context"
	"fmt"
	"github.com/tencentyun/cos-go-sdk-v5"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"skis-admin-backend/global"
	"time"
)

var ossClient *cos.Client

func InitOss() {

	u, _ := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", global.Config.Oss.Bucket, global.Config.Oss.EndPoint))
	b := &cos.BaseURL{BucketURL: u}
	ossClient = cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  global.Config.Oss.AccessKeyId,
			SecretKey: global.Config.Oss.AccessKeySecret,
		},
	})
}

func GetOssSign(ctx context.Context, key string) (string, error) {
	expire := int64(600)
	// 简单上传签名
	presignURL, err := ossClient.Object.GetPresignedURL(
		ctx,
		http.MethodPut,
		key,
		global.Config.Oss.AccessKeyId,
		global.Config.Oss.AccessKeySecret,
		time.Duration(expire)*time.Second,
		nil,
	)
	if err != nil {
		global.Lg.Error("获取oss签名失败", zap.Error(err))
		return "", err
	}
	return presignURL.String(), nil
}
