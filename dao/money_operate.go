package dao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	wechatUtils "github.com/wechatpay-apiv3/wechatpay-go/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
	"skis-admin-backend/services"
	"skis-admin-backend/services/transferbills"
	"time"
)

// MoneyOperateWithdraw 处理用户提现操作
// 参数:
//
//	c - gin上下文对象，包含用户认证信息
//	req - 提现请求参数，包含提现类型和金额
//
// 返回值:
//
//	error - 操作错误信息，nil表示成功
//
// 功能:
//  1. 验证用户类型(教练或俱乐部)
//  2. 检查用户当前提现中的总金额
//  3. 根据用户类型和提现类型(保证金/余额)验证账户余额是否充足
//  4. 创建提现操作记录
//
// 错误情况:
//   - 用户类型错误
//   - 教练/俱乐部信息查询失败
//   - 账户余额不足
//   - 数据库操作失败
func MoneyOperateWithdraw(c *gin.Context, req *forms.MoneyOperateWithdrawRequest) (resp *transferbills.CreateTransferBillResponse, err error) {
	userType := c.GetInt("user_type")
	if userType != enum.UserTypeCoach && userType != enum.UserTypeClub {
		return nil, errors.New("用户类型错误")
	}
	userId := c.GetString("user_id")

	//查询用户当前提现中的总金额
	op := model.MoneyOperate{}
	global.DB.Model(&model.MoneyOperate{}).
		Select("sum(money) as money").
		Where("user_id = ? and type = ? and status = ?", userId, req.Type, model.StatusWithdrawing).Scan(&op)

	var courseMoney int64
	if req.Type == model.TypeDeposit { //保证金提现时，要校验提现后的保证金是否大于最贵的课程金额
		course, _ := NewGoodsDao(c, global.DB).GetMaxPriceByUserId(userId)
		courseMoney = course.TeachMoney + course.AreaMoney
	}
	var openid, userName, uid, appid string
	if userType == enum.UserTypeCoach { //教练
		coach, err := CoachInfoByCoachId(userId)
		if err != nil {
			return nil, enum.NewErr(enum.CoachNotExistErr, "教练不存在")
		}
		if coach.FrozenDeposit == 1 {
			return nil, enum.NewErr(enum.CoachFrozenDepositErr, "保证金被冻结")
		}
		if req.Type == model.TypeDeposit { //提现保证金
			if coach.Deposit-op.Money-courseMoney < req.Money {
				return nil, enum.NewErr(enum.CoachDepositNotEnoughErr, "保证金不足")
			}
		} else if req.Type == model.TypeBalance { //提现余额
			if coach.Balance-op.Money < req.Money {
				return nil, enum.NewErr(enum.CoachBalanceNotEnoughErr, "余额不足")
			}
		}
		uid = coach.Uid
		userName = coach.Realname
		user, err := QueryUserInfo(uid)
		if err != nil || user == nil {
			return nil, err
		}
		openid = user.OpenId
		appid = global.Config.UserMiniProgram.AppId
	} else if userType == enum.UserTypeClub { //俱乐部
		club, err := QueryClubInfoByClubId(userId)
		if err != nil {
			return nil, enum.NewErr(enum.ClubExitErr, "俱乐部不存在")
		}
		if club.FrozenDeposit == 1 {
			return nil, enum.NewErr(enum.ClubFrozenDepositErr, "保证金被冻结")
		}
		if req.Type == model.TypeDeposit { //提现保证金
			if club.Deposit-op.Money-courseMoney < req.Money {
				return nil, enum.NewErr(enum.ClubDepositNotEnoughErr, "保证金不足")
			}
		} else if req.Type == model.TypeBalance { //提现余额
			if club.Balance-op.Money < req.Money {
				return nil, enum.NewErr(enum.ClubBalanceNotEnoughErr, "余额不足")
			}
		}
		userName = club.Manager
		uid = club.Uid
		user, err := QueryClubsUserInfo(uid)
		if err != nil || user == nil {
			return nil, err
		}
		openid = user.OpenId
		appid = global.Config.ClubMiniProgram.AppId
	}

	operateID := GenerateId("M")
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		//TODO 微信支付商家打钱给用户
		resp, err = Createtransferbill(c, operateID, appid, openid, userName, req.Money)
		if err != nil {
			global.Lg.Error("Createtransferbill err", zap.Error(err), zap.Any("req", req), zap.Any("resp", resp))
			return err
		}
		//创建提现操作记录
		operateData := model.MoneyOperate{
			OperateID:     operateID,
			UserID:        userId,
			UserType:      userType,
			Type:          req.Type,
			Money:         req.Money,
			Status:        model.StatusWithdrawing,
			OperateType:   model.OperateTypeWithdraw,
			TransactionID: *resp.TransferBillNo,
			PackageInfo:   *resp.PackageInfo,
		}
		err = global.DB.Create(&operateData).Error
		if err != nil {
			global.Lg.Error("提现失败, 保存失败", zap.Error(err), zap.Any("req", req), zap.Any("resp", resp), zap.Any("operateData", operateData))
			return err
		}
		updateData := map[string]interface{}{}
		if userType == enum.UserTypeCoach { //教练
			if req.Type == model.TypeDeposit { //提现保证金
				updateData["deposit"] = gorm.Expr("deposit - ?", req.Money)
			} else { //提现余额
				updateData["balance"] = gorm.Expr("balance - ?", req.Money)
			}
			err = tx.Table("coaches").Where("coach_id = ?", userId).Updates(updateData).Error
			if err != nil {
				global.Lg.Error("提现失败, 更新教练信息失败", zap.Error(err), zap.Any("req", req), zap.Any("resp", resp), zap.Any("operateData", operateData))
				return err
			}
		} else {
			if req.Type == model.TypeDeposit { //提现保证金
				updateData["deposit"] = gorm.Expr("deposit - ?", req.Money)
			} else { //提现余额
				updateData["balance"] = gorm.Expr("balance - ?", req.Money)
			}
			err = tx.Table("clubs").Where("club_id = ?", userId).Updates(updateData).Error
			if err != nil {
				global.Lg.Error("提现失败, 俱乐部信息更新失败", zap.Error(err), zap.Any("req", req), zap.Any("resp", resp), zap.Any("operateData", operateData))
				return err
			}
		}
		return err
	})
	return resp, nil
}

func Createtransferbill(ctx *gin.Context, outBillNo, appid, openid, userName string, transferAmount int64) (resp *transferbills.CreateTransferBillResponse, err error) {
	// 使用 utils 提供的函数从本地文件中加载商户私钥，商户私钥会用来生成请求的签名
	mchPrivateKey, err := wechatUtils.LoadPrivateKeyWithPath("./config/apiclient_key.pem")
	if err != nil {
		global.Lg.Error("load merchant private key error", zap.Error(err))
		return
	}

	wxPrivateKey, err := wechatUtils.LoadPublicKeyWithPath("./config/pub_key.pem")
	if err != nil {
		global.Lg.Error("load wxPay public key error", zap.Error(err))
		return
	}
	// 使用商户私钥等初始化 client，并使它具有自动定时获取微信支付平台证书的能力
	opts := []core.ClientOption{
		option.WithWechatPayPublicKeyAuthCipher(global.Config.Mch.MchId, global.Config.Mch.SerialNumber, mchPrivateKey, global.Config.Mch.PublicKeyId, wxPrivateKey),
	}
	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		global.Lg.Error("new wechat pay client err", zap.Error(err))
		return
	}

	if transferAmount <= 30 { //0.3元以下不允许输入姓名校验
		userName = ""
	}
	appId := ctx.GetString("app_id")
	svc := transferbills.TransferBillsApiService{Client: client}
	resp, result, err := svc.CreateTransferBill(ctx,
		transferbills.CreateTransferBillRequest{
			Appid:           core.String(appId),
			OutBillNo:       core.String(outBillNo),
			TransferSceneId: core.String("1005"), //佣金报酬
			Openid:          core.String(openid),
			TransferAmount:  core.Int64(transferAmount),
			TransferRemark:  core.String("用户提现"),
			TransferSceneReportInfos: []transferbills.TransferSceneReportInfo{
				{
					InfoType:    core.String("岗位类型"),
					InfoContent: core.String("E滑教练"),
				},
				{
					InfoType:    core.String("报酬说明"),
					InfoContent: core.String("奖励说明123"),
				},
			},
			UserName:           core.String(userName),
			NotifyUrl:          core.String(global.Config.Mch.TransferBillsNotifyUrl),
			UserRecvPerception: core.String("企业补贴"),
		},
	)

	if err != nil {
		// 处理错误
		global.Lg.Error("call CreateTransferBill err", zap.Error(err))
		if result.Response.StatusCode == 400 {
			str := fmt.Sprintf("%s", result.Response.Body)
			var billErr BillErr
			err = json.Unmarshal([]byte(str[1:len(str)-1]), &billErr)
			if err != nil {
				return nil, err
			}
			return nil, enum.NewErr(enum.WechatTransferBillErr, billErr.Code+":"+billErr.Message)
		}
	} else {
		// 处理返回结果
		global.Lg.Info("call CreateTransferBill", zap.Any("StatusCode", result.Response.StatusCode), zap.Any("resp", &resp))
	}
	return resp, err
}

type BillErr struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// MoneyOperateRecharge 处理用户充值请求
// 参数:
//
//	c - gin上下文对象
//	req - 充值请求表单数据
//
// 返回值:
//
//	error - 操作过程中发生的错误
//
// 功能:
//  1. 验证用户类型(教练或俱乐部)
//  2. 创建充值操作记录
//  3. 返回微信支付调用结果
func MoneyOperateRecharge(c *gin.Context, req *forms.MoneyOperateRechargeRequest) (*jsapi.PrepayWithRequestPaymentResponse, error) {
	userType := c.GetInt("user_type")
	if userType != enum.UserTypeCoach && userType != enum.UserTypeClub {
		return nil, errors.New("用户类型错误")
	}
	userId := c.GetString("user_id")

	operateData := model.MoneyOperate{
		OperateID: GenerateId("M"),
		UserID:    userId,
		UserType:  userType,
		Type:      model.TypeDeposit,
		Money:     req.Money,
		//Remark:   req.Remark,
		Status:      model.StatusRecharging,
		OperateType: model.OperateTypeRecharge,
	}
	err := global.DB.Create(&operateData).Error
	if err != nil {
		global.Lg.Error("充值失败", zap.Error(err), zap.Any("req", req))
		return nil, err
	}
	openId := c.GetString("open_id")
	uid := c.GetString("user_id")
	if userType == enum.UserTypeCoach {
		uid = c.GetString("coach_id")
	} else if userType == enum.UserTypeClub {
		uid = c.GetString("club_id")
	}
	//TODO 唤起微信支付
	appId := c.GetString("app_id")
	prePayResp, err := services.GeneratePreOrder(appId, uid, openId, operateData.OperateID, req.Money, "充值保证金："+fmt.Sprintf("%.2f", float64(req.Money)/100)+"元", global.Config.Mch.DepositPayNotifyUrl)
	if err != nil {
		global.Lg.Error("MoneyOperateRecharge: GeneratePreOrder error", zap.Error(err))
		return nil, err
	}
	return prePayResp, nil
}

// MoneyOperateRechargeCallback 保证金充值回调
func MoneyOperateRechargeCallback(c *gin.Context, r *http.Request) error {
	result, err := GetWxPayCallbackData(r)
	if err != nil {
		return err
	}
	global.Lg.Info("保证金充值回调", zap.Any("params", result))
	//这里只处理支付成功的情况
	//SUCCESS	支付成功	用户支付成功，资金已入账
	//REFUND	转入退款	交易已退款(全额或部分)
	//NOTPAY	未支付	订单已创建但未支付
	//CLOSED	已关闭	订单已关闭(商户或系统)
	//REVOKED	已撤销	付款码支付被用户撤销
	//USERPAYING	用户支付中	付款码支付用户已扫码但未确认
	//PAYERROR	支付失败	支付失败(余额不足等)
	if *result.TradeState != "SUCCESS" { //这里先只处理支付成功的情况
		global.Lg.Info("支付未成功", zap.Any("result", result))
		return nil
	}

	moneyOperate := model.MoneyOperate{}
	err = global.DB.Model(&model.MoneyOperate{}).Where("operate_id = ?", *result.OutTradeNo).First(&moneyOperate).Error
	if err != nil {
		global.Lg.Error("获取充值记录失败", zap.Error(err))
		return enum.NewErr(enum.MoneyOperateOrderNotExistErr, "保证金充值订单不存在") //找不到，返回成功
	}

	if moneyOperate.Status == model.StatusRecharged {
		global.Lg.Info("保证金充值订单订单已支付", zap.String("order_id", *result.OutTradeNo))
		return nil
	}

	payTime, err := time.Parse(time.RFC3339, *result.SuccessTime)
	if err != nil {
		global.Lg.Error("时间转换失败", zap.Error(err), zap.Any("result", result))
		payTime = time.Now()
	}

	err = global.DB.Transaction(func(tx *gorm.DB) error {
		//更新充值记录
		err = global.DB.Model(&model.MoneyOperate{}).Where("operate_id = ?", *result.OutTradeNo).Updates(map[string]interface{}{
			"status":         model.StatusRecharged,
			"transaction_id": *result.TransactionId,
			"pay_time":       payTime,
		}).Error
		if err != nil {
			global.Lg.Error("更新充值记录失败", zap.Error(err))
			return err
		}

		record := model.MoneyRecords{
			Money:        moneyOperate.Money,
			IncomeType:   model.IncomeTypeIncome,
			RelationType: model.RelationTypeDeposit,
			UserType:     moneyOperate.UserType,
			UserID:       moneyOperate.UserID,
			MoneyType:    model.CoachIncomeDepositRecharge,
			RelationID:   moneyOperate.OperateID,
		}
		err = NewMoneyRecordsDao(c, tx).Create(c, &record, tx)
		if err != nil {
			global.Lg.Error("创建资金流水失败", zap.Error(err), zap.Any("data", record), zap.Any("result", result))
			return err
		}
		return nil
	})
	global.Lg.Debug("订单支付成功", zap.Any("order", moneyOperate))

	return err
}

func TransferBillsCallback(c *gin.Context, r *http.Request) error {
	err := GetTransferBillsCallbackData(r)
	if err != nil {
		global.Lg.Error("转账回调失败", zap.Error(err), zap.Any("params", r))
		return err
	}

	return nil
}

type TransferBillsCallbackData struct {
	MchId          string    `json:"mch_id"`
	OutBillNo      string    `json:"out_bill_no"`
	TransferBillNo string    `json:"transfer_bill_no"`
	TransferAmount int       `json:"transfer_amount"`
	State          string    `json:"state"`
	Openid         string    `json:"openid"`
	CreateTime     time.Time `json:"create_time"`
	UpdateTime     time.Time `json:"update_time"`
	Mchid          string    `json:"mchid"`
}

func GetTransferBillsCallbackData(r *http.Request) (err error) {
	wechatpayPublicKey, err := wechatUtils.LoadPublicKeyWithPath("./config/pub_key.pem")
	if err != nil {
		global.Lg.Error("加载公钥失败", zap.Error(err))
		return
	}
	// 初始化 notify.Handler
	handler, err := notify.NewRSANotifyHandler(global.Config.Mch.ApiKey, verifiers.NewSHA256WithRSAPubkeyVerifier(global.Config.Mch.PublicKeyId, *wechatpayPublicKey))
	if err != nil {
		global.Lg.Error("创建回调处理器失败", zap.Error(err))
		return
	}

	transaction := new(payments.Transaction)
	notifyReq, err := handler.ParseNotifyRequest(context.Background(), r, transaction)
	// 如果验签未通过，或者解密失败
	if err != nil {
		global.Lg.Error("回调验签失败", zap.Error(err))
		return
	}
	//notifyReq.Resource.Plaintext = {"mch_id":"1726738364","out_bill_no":"M20251129105635deea76","transfer_bill_no":"1330008069721552511290045940397216","transfer_amount":30,"state":"SUCCESS","openid":"o4yE-1_aA_F-6M2AzthxsBOdhIRM","create_time":"2025-11-29T10:56:35+08:00","update_time":"2025-11-29T10:56:40+08:00","mchid":"1726738364"}
	global.Lg.Info("回调成功", zap.Any("result", notifyReq))
	// 解析通知内容为支付结果
	err = BillsCallback(notifyReq.Resource.Plaintext)
	if err != nil {
		global.Lg.Error("处理回调失败", zap.Error(err))
	}
	return err
}

func BillsCallback(plaintext string) (err error) {
	var result TransferBillsCallbackData
	if err = json.Unmarshal([]byte(plaintext), &result); err != nil {
		global.Lg.Error("解析回调内容失败", zap.Error(err))
		return
	}

	var moneyOperate model.MoneyOperate
	err = global.DB.Model(&model.MoneyOperate{}).Where("operate_id = ?", result.OutBillNo).Last(&moneyOperate).Error
	if err != nil {
		global.Lg.Error("获取提现记录失败", zap.Error(err), zap.Any("data", result))
		return
	}
	if moneyOperate.Status == model.StatusWithdrawed {
		global.Lg.Info("提现已处理", zap.String("order_id", result.OutBillNo))
		return
	}
	if moneyOperate.OperateType != model.OperateTypeWithdraw { //只处理提现的数据
		return
	}
	updateData := map[string]interface{}{}
	updateData["pay_log"] = plaintext
	if result.State == "SUCCESS" {
		updateData["status"] = model.StatusWithdrawed
		updateData["pay_time"] = result.UpdateTime
	} else {
		updateData["status"] = model.StatusWithdrawErr
	}
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(&model.MoneyOperate{}).Where("operate_id = ?", result.OutBillNo).Updates(updateData).Error
		if err != nil {
			global.Lg.Error("更新提现记录失败", zap.Error(err), zap.Any("data", updateData), zap.Any("result", result))
			return err
		}

		if result.State == "SUCCESS" {
			record := model.MoneyRecords{
				MoneyID:      GenerateId("ZJ"),
				Money:        moneyOperate.Money,
				IncomeType:   model.IncomeTypePay,
				RelationType: model.RelationTypeWithdraw,
				UserType:     moneyOperate.UserType,
				UserID:       moneyOperate.UserID,
				RelationID:   moneyOperate.OperateID,
			}
			if moneyOperate.UserType == enum.UserTypeCoach {
				if moneyOperate.Type == model.TypeDeposit {
					record.MoneyType = model.CoachPayDepositWithdraw
				} else {
					record.MoneyType = model.CoachPayFundsWithdraw
				}
			} else {
				if moneyOperate.Type == model.TypeDeposit {
					record.MoneyType = model.ClubPayDepositWithdraw
				} else {
					record.MoneyType = model.ClubPayFundsWithdraw
				}
			}
			record.MoneyDesc = model.UserMoneyTypeStr[record.MoneyType]
			record.Remark = model.UserMoneyTypeRemark[record.MoneyType]
			err = tx.Model(&model.MoneyRecords{}).Create(&record).Error
			if err != nil {
				global.Lg.Error("创建资金流水失败", zap.Error(err), zap.Any("data", record), zap.Any("result", result))
			}
			return err
		}
		//提现失败，需要将钱退回去
		money := moneyOperate.Money
		userId := moneyOperate.UserID
		if moneyOperate.UserType == enum.UserTypeCoach { //教练
			if moneyOperate.Type == model.TypeDeposit { //提现保证金
				updateData["deposit"] = gorm.Expr("deposit + ?", money)
			} else { //提现余额
				updateData["balance"] = gorm.Expr("balance + ?", money)
			}
			err = tx.Table("coaches").Where("coach_id = ?", userId).Updates(updateData).Error
			if err != nil {
				global.Lg.Error("提现失败, 更新教练信息失败", zap.Error(err), zap.Any("moneyOperate", moneyOperate))
				return err
			}
		} else {
			if moneyOperate.Type == model.TypeDeposit { //提现保证金
				updateData["deposit"] = gorm.Expr("deposit + ?", money)
			} else { //提现余额
				updateData["balance"] = gorm.Expr("balance + ?", money)
			}
			err = tx.Table("clubs").Where("club_id = ?", userId).Updates(updateData).Error
			if err != nil {
				global.Lg.Error("提现失败, 俱乐部信息更新失败", zap.Error(err), zap.Any("moneyOperate", moneyOperate))
				return err
			}
		}
		return nil
	})
	return err
}
