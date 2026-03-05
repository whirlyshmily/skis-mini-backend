package services

import (
	"context"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/transferbatch"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
	"log"
)

func main() {
	var (
		mchID                      string = "190000****"                               // 商户号
		mchCertificateSerialNumber string = "3775************************************" // 商户证书序列号
		mchAPIv3Key                string = "2ab9****************************"         // 商户APIv3密钥
	)

	// 使用 utils 提供的函数从本地文件中加载商户私钥，商户私钥会用来生成请求的签名
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath("/path/to/merchant/apiclient_key.pem")
	if err != nil {
		log.Printf("load merchant private key error:%s", err)
		return
	}

	ctx := context.Background()
	// 使用商户私钥等初始化 client，并使它具有自动定时获取微信支付平台证书的能力
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(mchID, mchCertificateSerialNumber, mchPrivateKey, mchAPIv3Key),
	}
	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		log.Printf("new wechat pay client err:%s", err)
		return
	}

	svc := transferbatch.TransferBatchApiService{Client: client}
	resp, result, err := svc.InitiateBatchTransfer(ctx,
		transferbatch.InitiateBatchTransferRequest{
			Appid:       core.String("wxf636efh567hg4356"),
			OutBatchNo:  core.String("plfk2020042013"),
			BatchName:   core.String("2019年1月深圳分部报销单"),
			BatchRemark: core.String("2019年1月深圳分部报销单"),
			TotalAmount: core.Int64(4000000),
			TotalNum:    core.Int64(200),
			TransferDetailList: []transferbatch.TransferDetailInput{transferbatch.TransferDetailInput{
				OutDetailNo:    core.String("x23zy545Bd5436"),
				TransferAmount: core.Int64(200000),
				TransferRemark: core.String("2020年4月报销"),
				Openid:         core.String("o-MYE42l80oelYMDE34nYD456Xoy"),
				UserName:       core.String("757b340b45ebef5467rter35gf464344v3542sdf4t6re4tb4f54ty45t4yyry45"),
			}},
			TransferSceneId: core.String("1000"),
		},
	)

	if err != nil {
		// 处理错误
		log.Printf("call InitiateBatchTransfer err:%s", err)
	} else {
		// 处理返回结果
		log.Printf("status=%d resp=%s", result.Response.StatusCode, resp)
	}
}
