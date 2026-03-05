package services

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"net/http"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/services/transferbills"
	"time"
)

type AccessToken struct {
	token    string
	expireAt int64
}

var AccessTokenCache *AccessToken

var WxPayClient *core.Client

func InitWxPayClient() error {
	// 使用 utils 提供的函数从本地文件中加载商户私钥，商户私钥会用来生成请求的签名
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath("./config/apiclient_key.pem")
	if err != nil {
		global.Lg.Error("load wxPay private key error", zap.Error(err))
		return err
	}

	wxPrivateKey, err := utils.LoadPublicKeyWithPath("./config/pub_key.pem")
	if err != nil {
		global.Lg.Error("load wxPay public key error", zap.Error(err))
		return err
	}

	ctx := context.Background()
	// 使用商户私钥等初始化 client，并使它具有自动定时获取微信支付平台证书的能力
	opts := []core.ClientOption{
		option.WithWechatPayPublicKeyAuthCipher(global.Config.Mch.MchId, global.Config.Mch.SerialNumber, mchPrivateKey, global.Config.Mch.PublicKeyId, wxPrivateKey),
	}
	WxPayClient, err = core.NewClient(ctx, opts...)
	if err != nil {
		global.Lg.Error("new wxPay client error", zap.Error(err))
		return err
	}
	return nil
}

func GeneratePreOrder(appId, userId, openId, orderId string, price int64, desc, notifyUrl string) (*jsapi.PrepayWithRequestPaymentResponse, error) {
	m := map[string]interface{}{
		"appId":     appId,
		"userId":    userId,
		"openId":    openId,
		"orderId":   orderId,
		"price":     price,
		"desc":      desc,
		"notifyUrl": notifyUrl,
	}
	global.Lg.Info("生成微信预订单", zap.Any("params", m))
	svc := jsapi.JsapiApiService{Client: WxPayClient}
	ctx := context.Background()
	resp, result, err := svc.PrepayWithRequestPayment(ctx,
		jsapi.PrepayRequest{
			Appid:       core.String(appId),
			Mchid:       core.String(global.Config.Mch.MchId),
			Description: core.String(desc),
			OutTradeNo:  core.String(orderId),
			Attach:      core.String("自定义数据说明"),
			NotifyUrl:   core.String(notifyUrl),
			Amount: &jsapi.Amount{
				Total: core.Int64(price),
			},
			Payer: &jsapi.Payer{
				Openid: core.String(openId),
			},
			SupportFapiao: core.Bool(true),
		},
	)

	if err != nil {
		global.Lg.Error("生成微信预订单失败", zap.Any("params", m))
		return nil, err
	}

	if result.Response.StatusCode != 200 {
		global.Lg.Error("生成微信预订单失败", zap.Any("params", m))
		return nil, fmt.Errorf("生成微信预订单失败: %v", result.Response.Status)
	}

	return resp, nil
}

func VerifyWechatCode(appId, secret, code string) (*forms.WechatLoginResponse, error) {
	// 从配置中获取小程序appid和secret
	global.Lg.Info("微信登录", zap.String("code", code), zap.String("appid", appId), zap.String("secret", secret))

	// 构造微信接口URL
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", appId, secret, code)

	// 发送请求
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var wechatResp forms.WechatLoginResponse
	err = json.Unmarshal(body, &wechatResp)
	if err != nil {
		global.Lg.Error("微信登录失败", zap.Error(err))
		return nil, err
	}

	// 检查微信返回的错误码
	if wechatResp.ErrCode != 0 {
		global.Lg.Error("微信登录失败", zap.Int("errcode", wechatResp.ErrCode), zap.String("errmsg", wechatResp.ErrMsg))
		return nil, fmt.Errorf("wechat login error: %d, %s", wechatResp.ErrCode, wechatResp.ErrMsg)
	}

	return &wechatResp, nil
}

// DecryptWechatPhoneData 解密微信手机号数据
func DecryptWechatPhoneData(sessionKey, iv, encryptedData string) (*forms.WechatPhoneInfo, error) {
	// base64解码
	sessionKeyBytes, err := base64.StdEncoding.DecodeString(sessionKey)
	if err != nil {
		return nil, err
	}

	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return nil, err
	}

	encryptedDataBytes, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}

	// AES解密
	block, err := aes.NewCipher(sessionKeyBytes)
	if err != nil {
		return nil, err
	}

	mode := cipher.NewCBCDecrypter(block, ivBytes)
	decryptedData := make([]byte, len(encryptedDataBytes))
	mode.CryptBlocks(decryptedData, encryptedDataBytes)

	// PKCS7解填充
	decryptedData = pkcs7Unpad(decryptedData, block.BlockSize())

	// 解析JSON
	var phoneInfo forms.WechatPhoneInfo
	err = json.Unmarshal(decryptedData, &phoneInfo)
	if err != nil {
		return nil, err
	}

	return &phoneInfo, nil
}

// PKCS7解填充
func pkcs7Unpad(data []byte, blockSize int) []byte {
	length := len(data)
	if length == 0 {
		return data
	}

	unpadding := int(data[length-1])
	if unpadding >= length || unpadding > blockSize {
		return data
	}

	return data[:(length - unpadding)]
}

func RefundOrder(orderId, transactionId, outRefundNo string, amount int64, notifyUrl string) (*refunddomestic.Refund, error) {
	// 构建退款请求参数
	req := refunddomestic.CreateRequest{
		TransactionId: core.String(transactionId), // 微信支付订单号（与商户订单号二选一）
		OutTradeNo:    core.String(orderId),       // 商户订单号（与微信支付订单号二选一）
		OutRefundNo:   core.String(outRefundNo),
		Reason:        core.String("用户主动退款"), // 退款原因（可选）
		Amount: &refunddomestic.AmountReq{
			Total:    core.Int64(amount), // 原订单总金额（分）
			Refund:   core.Int64(amount), // 退款金额（分）
			Currency: core.String("CNY"),
		},
		NotifyUrl: core.String(notifyUrl), // 退款结果通知地址
	}

	svc := refunddomestic.RefundsApiService{
		Client: WxPayClient,
	}
	resp, result, err := svc.Create(context.Background(), req)
	if err != nil {
		global.Lg.Error("申请退款失败", zap.Error(err), zap.Any("req", req))
		return nil, err
	}

	if result.Response.StatusCode != 200 {
		global.Lg.Error("申请退款失败", zap.Any("req", req))
		return nil, fmt.Errorf("申请退款失败: %v", result.Response.Status)
	}

	return resp, nil
}

func CreateTransferBill(Appid, OutBillNo, TransferSceneId, Openid string, TransferAmount int64, TransferRemark, UserName, NotifyUrl, UserRecvPerception string) (resp *transferbills.CreateTransferBillResponse, err error) {
	ctx := context.Background()
	svc := transferbills.TransferBillsApiService{Client: WxPayClient}
	tr := transferbills.CreateTransferBillRequest{
		Appid:           core.String(Appid),
		OutBillNo:       core.String(OutBillNo),
		TransferSceneId: core.String(TransferSceneId),
		Openid:          core.String(Openid),
		TransferAmount:  core.Int64(TransferAmount),
		TransferRemark:  core.String(TransferRemark),
		TransferSceneReportInfos: []transferbills.TransferSceneReportInfo{
			{
				InfoType:    core.String("岗位类型"),
				InfoContent: core.String("教练"),
			}, {
				InfoType:    core.String("报酬说明"),
				InfoContent: core.String("7月报酬"),
			},
		},
		UserName:  core.String(UserName),
		NotifyUrl: core.String(NotifyUrl),
		//UserRecvPerception: core.String(UserRecvPerception),
	}
	resp, result, err := svc.CreateTransferBill(ctx, tr)

	//global.Lg.Info("商家转账", zap.Any("resp", resp), zap.Any("result", result), zap.Any("tr", tr))
	if err != nil {
		// 处理错误
		global.Lg.Error("call CreateTransferBill err:%s", zap.Error(err))
		log.Printf("status=%d resp=%s", result.Response.StatusCode, resp)
		return nil, err
	}

	if result.Response.StatusCode != 200 {
		global.Lg.Error("申请退款失败", zap.Any("req", resp))
		return nil, fmt.Errorf("申请退款失败: %v", result.Response.Status)
	}
	return resp, err
}

type GetAccessTokenResponse struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func GetAccessToken() (*GetAccessTokenResponse, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", global.Config.UserMiniProgram.AppId, global.Config.UserMiniProgram.Secret)
	resp, err := http.Get(url)
	if err != nil {
		global.Lg.Error("获取access_token失败", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		global.Lg.Error("获取access_token失败", zap.Error(err))
		return nil, err
	}
	var wechatResp GetAccessTokenResponse
	err = json.Unmarshal(body, &wechatResp)
	if err != nil {
		global.Lg.Error("获取access_token失败", zap.Error(err))
		return nil, err
	}
	return &wechatResp, nil
}

func getAccessToken() (string, error) {
	if AccessTokenCache == nil || AccessTokenCache.expireAt < time.Now().Unix() {
		getAccessTokenRsp, err := GetAccessToken()
		if err != nil {
			global.Lg.Error("获取access_token失败", zap.Error(err))
			return "", err
		}

		AccessTokenCache = &AccessToken{
			token:    getAccessTokenRsp.AccessToken,
			expireAt: time.Now().Unix() + int64(getAccessTokenRsp.ExpiresIn),
		}
	}

	return AccessTokenCache.token, nil
}

type GetUserPhoneInfoRequest struct {
	Code string `json:"code"`
}

func GetUserPhoneInfo(ctx context.Context, code string) (*forms.WechatPhoneInfo, error) {
	token, err := getAccessToken()
	if err != nil {
		global.Lg.Error("getAccessToken failed:%v", zap.Error(err))
		return nil, err
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s", token)
	req := GetUserPhoneInfoRequest{
		Code: code,
	}

	params, err := json.Marshal(req)
	if err != nil {
		global.Lg.Error("json.Marshal failed:%v", zap.Error(err))
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(params))
	if err != nil {
		global.Lg.Error("http.Post failed:%v", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		global.Lg.Error("ioutil.ReadAll failed:%v", zap.Error(err))
		return nil, err
	}

	var wechatResp forms.WechatPhoneInfo
	err = json.Unmarshal(body, &wechatResp)
	if err != nil {
		global.Lg.Error("json.Unmarshal failed:%v", zap.Error(err))
		return nil, err
	}

	if wechatResp.Errcode != 0 {
		global.Lg.Error("get user phone info failed", zap.Any("wechatResp", wechatResp))
		return nil, fmt.Errorf("get user phone info failed: %s", wechatResp.Errcode)
	}

	return &wechatResp, nil
}
