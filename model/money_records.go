package model

import "time"

//create table money_records
//(
//id            bigint auto_increment comment '自增ID'
//primary key,
//money_id      varchar(64)                           null comment '资金id',
//user_id       varchar(64)                           not null comment '用户ID（包括用户、教练、俱乐部、E滑官方）',
//user_type     int         default 1                 not null comment '用户类型（1：用户、2：教练、3：俱乐部、4：E滑官方）',
//money         int         default 0                 not null comment '资金变化（单位：分）',
//money_type    int         default 0                 not null comment '资金类型（53种，https://lanhuapp.com/web/#/item/project/product?pid=6982fcff-f9f4-4f54-9b72-379b8d22866a&teamId=1d91bc4a-48fe-4734-b0c7-b165b09dcdc5&tid=1d91bc4a-48fe-4734-b0c7-b165b09dcdc5&versionId=b2788027-a439-44f7-9572-817febdc7ff8&docId=bcfc8462-4c73-422c-9f5f-70ad9173f9bf&docType=axure&pageId=0f468d8af9c34e26bb0ad3d75bae784c&image_id=bcfc8462-4c73-422c-9f5f-70ad9173f9bf&parentId=6867bff1-00a8-46ec-af71-e64b4556cd2a）',
//income_type   int         default 0                 not null comment '收入类型（0：收入，1：支出）',
//relation_id   varchar(64)                           null comment '关联id',
//relation_type int                                   null comment '产生对应的表（0：订单，1：保证金）',
//order_course_id varchar(64)                           null comment '订单课程ID',
//money_desc    varchar(64) default ''                not null comment '资金流水描述',
//remark        varchar(255)                          null comment '备注',
//state         int         default 0                 not null comment '状态（0：正常，1：删除）',
//created_at    timestamp   default CURRENT_TIMESTAMP not null comment '创建时间',
//updated_at    timestamp   default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP comment '更新时间'
//)
//comment '产生资金流水环节 ' row_format = DYNAMIC;

type MoneyRecords struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	UserID        string    `json:"user_id" gorm:"column:user_id"`                 // 用户ID（包括用户、教练、俱乐部、E滑官方）
	MoneyID       string    `json:"money_id" gorm:"column:money_id"`               // 资金ID
	UserType      int       `json:"user_type" gorm:"column:user_type"`             // 用户类型（1：用户、2：教练、3：俱乐部、4：E滑官方）
	Money         int64     `json:"money" gorm:"column:money"`                     // 资金变化（单位：分）
	MoneyType     int       `json:"money_type" gorm:"column:money_type"`           // 资金类型（53种，https://lanhuapp.com/web/#/item/project/product?pid=6982fcff-f9f4-4f54-9b72-379b8d22866a&teamId=1d91bc4a-48fe-4734-b0c7-b165b09dcdc5&tid=1d91bc4a-48fe-4734-b0c7-b165b09dcdc5&versionId=b2788027-a439-44f7-9572-817febdc7ff8&docId=bcfc8462-4c73-422c-9f5f-70ad9173f9bf&docType=axure&pageId=0f468d8af9c34e26bb0ad3d75bae784c&image_id=bcfc8462-4c73-422c-9f5f-70ad9173f9bf&parentId=6867bff1-00a8-46ec-af71-e64b4556cd2a）
	IncomeType    int       `json:"income_type" gorm:"column:income_type"`         // 收入类型（0：收入，1：支出）
	RelationID    string    `json:"relation_id" gorm:"column:relation_id"`         // 关联id
	RelationType  int       `json:"relation_type" gorm:"column:relation_type"`     // 产生对应的表（0：订单，1：保证金）
	OrderCourseID string    `json:"order_course_id" gorm:"column:order_course_id"` // 订单课程ID
	MoneyDesc     string    `json:"money_desc" gorm:"column:money_desc"`           // 资金流水描述
	Remark        string    `json:"remark" gorm:"column:remark"`                   // 备注
	State         int       `json:"-" gorm:"column:state"`                         // 状态（0：正常，1：删除）
	CreatedAt     LocalTime `json:"created_at" gorm:"column:created_at"`           // 创建时间
	UpdatedAt     time.Time `json:"updated_at" gorm:"column:updated_at"`           // 更新时间
}

func (m *MoneyRecords) TableName() string {
	return "money_records"
}

// 资金类型53种，https://lanhuapp.com/web/#/item/project/product?pid=6982fcff-f9f4-4f54-9b72-379b8d22866a&teamId=1d91bc4a-48fe-4734-b0c7-b165b09dcdc5&tid=1d91bc4a-48fe-4734-b0c7-b165b09dcdc5&versionId=b2788027-a439-44f7-9572-817febdc7ff8&docId=bcfc8462-4c73-422c-9f5f-70ad9173f9bf&docType=axure&pageId=0f468d8af9c34e26bb0ad3d75bae784c&image_id=bcfc8462-4c73-422c-9f5f-70ad9173f9bf&parentId=6867bff1-00a8-46ec-af71-e64b4556cd2a
// https://docs.qq.com/sheet/DSnB1cUpSYXRoUm1n?tab=000001
const (
	// 用户支出类型（1）
	UserPayBuyCourse = 1001 // 用户购买课程

	// 用户收入类型（11）
	UserIncomeRefundNoFault                 = 2002 //用户退课（无责）课程实付金额
	UserIncomeOneCRefundFault               = 2003 //单次课程退课 （有责）课程实付金额 扣除有责取消费
	UserIncomeOneCTranferRefundFault        = 2006 //单次课程退课 （转单、有责）课程实付金额（扣除有责取消费）
	UserIncomePackCRefundNoFinish           = 2009 //打包课程退课（无已完成课程）课程实付金额
	UserIncomePackCRefundFinishToClub       = 2010 //打包课程退课（含已完成课程、俱乐部课程）课程实付金额（扣除所有已完成课程原价合计）
	UserIncomePackCRefundFinishToCoach      = 2015 //打包课程退课（含已完成课程、教练课程）课程实付金额（扣除所有已完成课程原价合计）
	UserIncomeAdminRefund                   = 2037 //后台退课（全额退款）课程实付金额
	UserIncomeAdminPartRefund               = 2038 //后台退课（部分退款，无转单）后台设置学员退课金额
	UserIncomeAdminPartRefundToTranfer      = 2041 //后台退课（部分退款，转单）后台设置学员退课金额
	UserIncomeAdminPartRefundToClubNoCoach  = 2046 //后台退课（俱乐部，部分退款，无教练接课）后台设置学员退课金额
	UserIncomeAdminPartRefundToClubYesCoach = 2049 //后台退课（俱乐部，部分退款，有教练接课）后台设置学员退课金额

	// 教练支出类型（13）
	CoachPayOneCRefundFaultService                  = 3005 //单次课退课服务费
	CoachPayOneCTranferRefundFaultService           = 3008 //单次课程退课 （转单、有责）平台服务费（接单教练）
	CoachPayPackCRefundFinishToClubService          = 3012 //打包课程退课（含已完成课程、俱乐部课程）平台服务费（已完成课程原价减去打包课程价差价（不含推荐费））
	CoachPayPackCRefundFinishToCoachService         = 3017 //打包课程退课（含已完成课程、教练课程）平台服务费（已完成课程原价减去打包课程价差价）
	CoachPayFinishCToCoachService                   = 3019 //完成课程（教练）平台服务费
	CoachPayFinishCToTranferService                 = 3022 //完成课程（转单）平台服务费
	CoachPayFinishCToTranferBySellService           = 3024 //完成课程（转单）平台服务费（原教练）
	CoachPayFinishCToClubService                    = 3027 //完成俱乐部课程平台服务费
	CoachPayDepositWithdraw                         = 3033 //保证金提取
	CoachPayFundsWithdraw                           = 3035 //钱包资金提取 收取资金
	CoachPayDepositWithdrawRefund                   = 3038 //保证金提现失败退回
	CoachPayBalanceWithdrawRefund                   = 3039 //余额提现失败退回
	CoachPayAdminPartRefundService                  = 3040 //后台退课（部分退款，无转单）平台服务费（教练收到退款）
	CoachPayAdminPartRefundToTranferBySellerService = 3043 //后台退课（部分退款，转单）平台服务费（原教练收到退款）
	CoachPayAdminPartRefundToTranferService         = 3045 //后台退课（部分退款，转单）平台服务费（接单教练收到退款）
	CoachPayAdminPartRefundToClubYesCoachService    = 3053 //后台退课（俱乐部，部分退款，有教练接课）平台服务费（接单教练收到退款）
	// CoachPayDepositRecharge                         = 3031 //教练充值保证金

	// 教练收入类型（17）
	CoachIncomeOneCRefundFault                  = 4004 //单次课程退课 （有责）有责取消费
	CoachIncomeOneCTranferRefundFault           = 4007 //单次课程退课 （转单、有责）有责取消费（接单教练）
	CoachIncomePackCRefundFinishToClub          = 4011 //打包课程退课（含已完成课程、俱乐部课程）已完成课程原价减去打包课程价差价（不含推荐费）
	CoachIncomePackCRefundFinishToCoach         = 4016 //打包课程退课（含已完成课程、教练课程）已完成课程原价减去打包课程价差价
	CoachIncomeFinishCToCoach                   = 4018 //完成课程（教练）课程金额
	CoachIncomeFinishCToCoachReferralBonus      = 4020 //完成课程（教练）推荐关系奖励
	CoachIncomeFinishCToTranferByTeacher        = 4021 //完成课程（转单）课程转单金额（接单教练）
	CoachIncomeFinishCToTranferBySeller         = 4023 //完成课程（转单）课程差价（原教练）
	CoachIncomeFinishCToTranferReferralBonus    = 4025 //完成课程（转单）推荐关系奖励
	CoachIncomeFinishCToClub                    = 4026 //完成课程（俱乐部）课程金额
	CoachIncomeFinishCToClubReferralBonus       = 4030 //完成课程（俱乐部）推荐关系奖励
	CoachIncomeDepositRecharge                  = 4031 //教练充值保证金
	CoachIncomeAdminPartRefund                  = 4039 //后台退课（部分退款，无转单）退课金额（实付金额扣除后台设置学员退课金额）
	CoachIncomeAdminPartRefundToTranferBySeller = 4042 //后台退课（部分退款，转单）原教练退款（实付金额扣除学员、接单教练退课金额）
	CoachIncomeAdminPartRefundToTranfer         = 4044 //后台退课（部分退款，转单）接单教练退课金额
	CoachIncomeAdminPartRefundToClubYesCoach    = 4052 //后台退课（俱乐部，部分退款，有教练接课）接单教练退课金额
	// CoachIncomeDepositWithdraw               = 4033 //保证金提取
	//CoachIncomeFundsWithdraw                 = 4035 //钱包资金提取 收取资金

	// 俱乐部支出类型（5）
	ClubPayPackCRefundFinishToClubService       = 5014 //打包课程退课（含已完成课程、俱乐部课程）平台服务费（已完成课程推荐费原价减去打包推荐费差价）
	ClubPayFinishCToClubService                 = 5029 //完成课程（俱乐部）平台服务费
	ClubPayAdminPartRefundToClubNoCoachService  = 5048 //后台退课（俱乐部，部分退款，无教练接课）平台服务费（俱乐部收到退款）
	ClubPayAdminPartRefundToClubYesCoachService = 5051 //后台退课（俱乐部，部分退款，有教练接课）平台服务费（俱乐部收到退款）
	ClubPayDepositWithdraw                      = 5034 //保证金提取
	ClubPayFundsWithdraw                        = 5036 //提取资金
	ClubPayDepositWithdrawRefund                = 5039 //保证金提现失败退回
	ClubPayBalanceWithdrawRefund                = 5040 //余额提现失败退回
	//ClubPayDepositRecharge                      = 5032 //俱乐部充值保证金

	// 俱乐部收入类型（6）
	ClubIncomePackCRefundFinishToClub       = 6013 //打包课程退课（含已完成课程、俱乐部课程） 已完成课程推荐费原价减去打包推荐费差价
	ClubIncomeFinishCToClub                 = 6028 //完成课程（俱乐部）推荐关系奖励
	ClubIncomeDepositRecharge               = 6032 //俱乐部充值保证金
	ClubIncomeAdminPartRefundToClubNoCoach  = 6047 //后台退课（俱乐部，部分退款，无教练接课）课程实付金额（扣除后台设置学员退课金额）
	ClubIncomeAdminPartRefundToClubYesCoach = 6050 //后台退课（俱乐部，部分退款，有教练接课）课程实付金额（扣除后台设置学员、接单教练退课金额）
	//ClubIncomeDepositWithdraw               = 6034 //保证金提取
	//ClubIncomeFundsWithdraw                 = 6036 //收取资金
)

var UserMoneyTypeStr = map[int]string{
	UserPayBuyCourse:                                "购买课程",
	UserIncomeRefundNoFault:                         "课程退课全额退款",
	UserIncomeOneCRefundFault:                       "课程退课剩余退款",
	CoachIncomeOneCRefundFault:                      "课程有责取消退款",
	CoachPayOneCRefundFaultService:                  "平台服务费",
	UserIncomeOneCTranferRefundFault:                "课程退课剩余退款",
	CoachIncomeOneCTranferRefundFault:               "课程有责取消退款",
	CoachPayOneCTranferRefundFaultService:           "平台服务费",
	UserIncomePackCRefundNoFinish:                   "课程退课全额退款",
	UserIncomePackCRefundFinishToClub:               "课程退课剩余退款",
	CoachIncomePackCRefundFinishToClub:              "打包课程退课剩余退款",
	CoachPayPackCRefundFinishToClubService:          "平台服务费",
	ClubIncomePackCRefundFinishToClub:               "打包课程退课剩余退款",
	ClubPayPackCRefundFinishToClubService:           "平台服务费",
	UserIncomePackCRefundFinishToCoach:              "打包课程退课剩余退款",
	CoachIncomePackCRefundFinishToCoach:             "打包课程退课剩余退款",
	CoachPayPackCRefundFinishToCoachService:         "平台服务费",
	CoachIncomeFinishCToCoach:                       "课程收入",
	CoachPayFinishCToCoachService:                   "平台服务费",
	CoachIncomeFinishCToCoachReferralBonus:          "推荐关系奖励收入",
	CoachIncomeFinishCToTranferByTeacher:            "转单课程收入",
	CoachPayFinishCToTranferService:                 "平台服务费",
	CoachIncomeFinishCToTranferBySeller:             "课程转单差价收入",
	CoachPayFinishCToTranferBySellService:           "平台服务费",
	CoachIncomeFinishCToTranferReferralBonus:        "推荐关系奖励收入",
	CoachIncomeFinishCToClub:                        "课程收入",
	CoachPayFinishCToClubService:                    "平台服务费",
	ClubIncomeFinishCToClub:                         "课程推荐收入",
	ClubPayFinishCToClubService:                     "平台服务费",
	CoachIncomeFinishCToClubReferralBonus:           "推荐关系奖励收入",
	CoachIncomeDepositRecharge:                      "充值保证金",
	ClubIncomeDepositRecharge:                       "充值保证金",
	CoachPayDepositWithdraw:                         "提取保证金",
	ClubPayDepositWithdraw:                          "提取保证金",
	CoachPayFundsWithdraw:                           "提取钱包资金",
	ClubPayFundsWithdraw:                            "提取钱包资金",
	CoachPayDepositWithdrawRefund:                   "保证金提现失败退回",
	CoachPayBalanceWithdrawRefund:                   "余额提现失败退回",
	ClubPayDepositWithdrawRefund:                    "保证金提现失败退回",
	ClubPayBalanceWithdrawRefund:                    "余额提现失败退回",
	UserIncomeAdminRefund:                           "客服系统全额退单退款",
	UserIncomeAdminPartRefund:                       "客服系统协商退单退款",
	CoachIncomeAdminPartRefund:                      "客服系统协商退单退款",
	CoachPayAdminPartRefundService:                  "平台服务费",
	UserIncomeAdminPartRefundToTranfer:              "客服系统协商退单退款",
	CoachIncomeAdminPartRefundToTranferBySeller:     "客服系统协商退单退款",
	CoachPayAdminPartRefundToTranferBySellerService: "平台服务费",
	CoachIncomeAdminPartRefundToTranfer:             "客服系统协商退单退款",
	CoachPayAdminPartRefundToTranferService:         "平台服务费",
	UserIncomeAdminPartRefundToClubNoCoach:          "客服系统协商退单退款",
	ClubIncomeAdminPartRefundToClubNoCoach:          "客服系统协商退单退款",
	ClubPayAdminPartRefundToClubNoCoachService:      "平台服务费",
	UserIncomeAdminPartRefundToClubYesCoach:         "客服系统协商退单退款",
	ClubIncomeAdminPartRefundToClubYesCoach:         "客服系统协商退单退款",
	ClubPayAdminPartRefundToClubYesCoachService:     "平台服务费",
	CoachIncomeAdminPartRefundToClubYesCoach:        "客服系统协商退单退款",
	CoachPayAdminPartRefundToClubYesCoachService:    "平台服务费",
}

var UserMoneyTypeRemark = map[int]string{
	UserPayBuyCourse:                                "用户微信支付至E滑微信账户",
	UserIncomeRefundNoFault:                         "E滑微信账户至用户微信支付",
	UserIncomeOneCRefundFault:                       "E滑微信账户至用户微信支付",
	CoachIncomeOneCRefundFault:                      "平台内部流转",
	CoachPayOneCRefundFaultService:                  "平台内部流转",
	UserIncomeOneCTranferRefundFault:                "E滑微信账户至用户微信支付",
	CoachIncomeOneCTranferRefundFault:               "平台内部流转",
	CoachPayOneCTranferRefundFaultService:           "平台内部流转",
	UserIncomePackCRefundNoFinish:                   "E滑微信账户至用户微信支付",
	UserIncomePackCRefundFinishToClub:               "E滑微信账户至用户微信支付",
	CoachIncomePackCRefundFinishToClub:              "平台内部流转",
	CoachPayPackCRefundFinishToClubService:          "平台内部流转",
	ClubIncomePackCRefundFinishToClub:               "平台内部流转",
	ClubPayPackCRefundFinishToClubService:           "平台内部流转",
	UserIncomePackCRefundFinishToCoach:              "E滑微信账户至用户微信支付",
	CoachIncomePackCRefundFinishToCoach:             "平台内部流转",
	CoachPayPackCRefundFinishToCoachService:         "平台内部流转",
	CoachIncomeFinishCToCoach:                       "平台内部流转",
	CoachPayFinishCToCoachService:                   "平台内部流转",
	CoachIncomeFinishCToCoachReferralBonus:          "平台内部流转",
	CoachIncomeFinishCToTranferByTeacher:            "平台内部流转",
	CoachPayFinishCToTranferService:                 "平台内部流转",
	CoachIncomeFinishCToTranferBySeller:             "平台内部流转",
	CoachPayFinishCToTranferBySellService:           "平台内部流转",
	CoachIncomeFinishCToTranferReferralBonus:        "平台内部流转",
	CoachIncomeFinishCToClub:                        "平台内部流转",
	CoachPayFinishCToClubService:                    "平台内部流转",
	ClubIncomeFinishCToClub:                         "平台内部流转",
	ClubPayFinishCToClubService:                     "平台内部流转",
	CoachIncomeFinishCToClubReferralBonus:           "平台内部流转",
	CoachIncomeDepositRecharge:                      "用户微信支付至E滑微信账户",
	ClubIncomeDepositRecharge:                       "用户微信支付至E滑微信账户",
	CoachPayDepositWithdraw:                         "E滑微信账户至用户微信支付",
	ClubPayDepositWithdraw:                          "E滑微信账户至用户微信支付",
	CoachPayFundsWithdraw:                           "E滑微信账户至用户微信支付",
	ClubPayFundsWithdraw:                            "E滑微信账户至用户微信支付",
	UserIncomeAdminRefund:                           "E滑微信账户至用户微信支付",
	UserIncomeAdminPartRefund:                       "E滑微信账户至用户微信支付",
	CoachPayDepositWithdrawRefund:                   "平台内部流转（提现失败退回）",
	CoachPayBalanceWithdrawRefund:                   "平台内部流转（提现失败退回）",
	ClubPayDepositWithdrawRefund:                    "平台内部流转（提现失败退回）",
	ClubPayBalanceWithdrawRefund:                    "平台内部流转（提现失败退回）",
	CoachIncomeAdminPartRefund:                      "平台内部流转",
	CoachPayAdminPartRefundService:                  "平台内部流转",
	UserIncomeAdminPartRefundToTranfer:              "E滑微信账户至用户微信支付",
	CoachIncomeAdminPartRefundToTranferBySeller:     "平台内部流转",
	CoachPayAdminPartRefundToTranferBySellerService: "平台内部流转",
	CoachIncomeAdminPartRefundToTranfer:             "平台内部流转",
	CoachPayAdminPartRefundToTranferService:         "平台内部流转",
	UserIncomeAdminPartRefundToClubNoCoach:          "E滑微信账户至用户微信支付",
	ClubIncomeAdminPartRefundToClubNoCoach:          "平台内部流转",
	ClubPayAdminPartRefundToClubNoCoachService:      "平台内部流转",
	UserIncomeAdminPartRefundToClubYesCoach:         "E滑微信账户至用户微信支付",
	ClubIncomeAdminPartRefundToClubYesCoach:         "平台内部流转",
	ClubPayAdminPartRefundToClubYesCoachService:     "平台内部流转",
	CoachIncomeAdminPartRefundToClubYesCoach:        "平台内部流转",
	CoachPayAdminPartRefundToClubYesCoachService:    "平台内部流转",
}

const RelationTypeOrder = 0    // 关联类型：订单
const RelationTypeDeposit = 1  // 关联类型：保证金
const RelationTypeWithdraw = 2 // 关联类型：提现

const IncomeTypeIncome = 0 // 收入类型：收入
const IncomeTypePay = 1    // 收入类型：支出（相对于钱包余额来说）
