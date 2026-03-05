package enum

const (
	//通用错误码
	Success         = 0   //成功
	GeneralErr      = 100 //通用错误
	ParamErr        = 101 //参数错误
	TokenInvalidErr = 102 //token无效
	DBErr           = 103 //数据库错误
	TeachTimeErr    = 104 //时间格式错误
	UserTypeError   = 105 //0-用户类型错误
	BadRequestErr   = 106 //请求参数错误
	AuthErr         = 107 //权限错误

	//业务错误码
	UsernameOrPwdErr  = 1001 //用户名或密码错误
	UsernameExistErr  = 1002 //用户名已存在
	DataNotExist      = 1003 //数据不存在
	RoleNotExistErr   = 1004 //角色不存在
	PasswordErr       = 1005 //密码错误
	SuperAdminErr     = 1006 //不能删除超管角色
	RoleExistUserErr  = 1007 //角色下有用户，不能删除
	RightExistRoleErr = 1008 //权限下有角色，不能删除
	TagExistErr       = 1009 //标签已存在
	CourseRefTagErr   = 1010 //课程已被引用
	CoachRefTagErr    = 1011 //教练已被引用
	SkiResortExistErr = 1012 //滑雪场已存在
	UserCreditsErr    = 1013 //用户积分不足

	GoodNoExistErr   = 1100 //商品不存在
	GoodNoStackErr   = 1101 //商品库存不足
	GoodNoCoursesErr = 1102 //商品下无课程

	CoachCourseTagExistErr  = 1200 //教练标签已存在
	CoachNotExistErr        = 1201 //教练不存在
	CoachDelTagErr          = 1202 //教练删除失败
	CoachTagNoExistErr      = 1203 //教练技能不存在
	CoachCreateSkiErr       = 1204 //场地插入失败
	CoachUpdateSkiErr       = 1205 //场地更新失败
	CoachDepositLackErr     = 1206 //教练保证金不足
	CoachFrozenDepositErr   = 1207 //教练保证金冻结失败
	CoachUnFrozenDepositErr = 1208 //教练保证金解冻失败
	CoachSkiResortErr       = 1209 //教练场地不存在
	CoachVerifyErr          = 1210 //教练审核失败
	CoachHasVerifiedErr     = 1211 //教练已通过认证

	SkiResortsGetErr = 1300 //场地查询失败

	CoachTagReviewCreateErr = 1400 //技能申请插入失败
	CoachTagReviewGetErr    = 1401 //技能申请查询失败

	CoachCertificateReviewCreateErr = 1500 //证书申请插入失败
	CertificateLevelErr             = 1501 //证书级别错误
	CoachCertificateReviewGetErr    = 1502 //证书查询失败
	CertificateExitErr              = 1503 //证书查询失败

	//俱乐部
	ClubExitErr            = 1600 //俱乐部不存在
	CoachQuitErr           = 1601 //退出俱乐部失败
	ClubJoinErr            = 1602 //加入俱乐部失败
	ClubDepositLackErr     = 1603 //俱乐部保证金不足
	ClubCoachJoinErr       = 1604 //教练加入俱乐部失败
	ClubCoachQuitErr       = 1605 //教练退出俱乐部失败
	ClubUpdateErr          = 1606 //俱乐部信息更新失败
	ClubFrozenDepositErr   = 1607 //俱乐部保证金被冻结
	ClubUnFrozenDepositErr = 1608 //俱乐部保证金解冻失败
	ClubCoachCheckErr      = 1609 //俱乐部审核教练失败

	//课程
	CourseExitErr = 1700 //课程不存在

	//商品
	GoodsExitErr        = 1800 //商品不存在
	GoodsNoStockErr     = 1801 //商品库存不足
	GoodsNoExistErr     = 1802 //商品不存在
	GoodsNoEnoughErr    = 1803 //商品库存不足
	GoodsAddErr         = 1804 //课程添加失败
	GoodsCoursesAddErr  = 1805 //商品课程关联添加失败
	GoodsTeachTimeErr   = 1806 //单次课程时长必须在60-480分钟
	GoodsTeachMoneyErr  = 1807 //单次课程价格必须在1-9999元
	GoodsAreaMoneyErr   = 1808 //单次课程场地费用必须在1-9999元
	GoodsFaultMoneyErr  = 1809 //单次课程场地费用必须在1元-场地费用
	GoodsTakeDownErr    = 1810 //商品下架失败
	GoodsPutUpErr       = 1811 //商品上架失败
	GoodsDeleteErr      = 1812 //商品删除失败
	GoodsEditErr        = 1813 //商品编辑失败
	GoodsCoursesEditErr = 1814 //商品课程编辑失败
	GoodsNotExistErr    = 1815 //商品不存在
	GoodsPackErr        = 1816 //商品打包失败
	GoodsUnPackErr      = 1817 //商品解包失败

	// 订单业务错误码
	OrderNotExistErr                 = 1900 //订单不存在
	OrdersCoursesExitErr             = 1920 //存在未核销的课程
	OrdersCreateCreditsRecordErr     = 1921 //创建积分记录失败
	OrderUnpaidErr                   = 1922 //订单未支付
	OrderCourseRefundExistErr        = 1923 //订单课程已申请退款
	OrderCourseFinishedOrCanceledErr = 1924 //订单课程已结束或已取消
	OrderFrozenMoneyErr              = 1925 //订单冻结
	OrderCourseNoRefundErr           = 1926 //订单无需退款

	//流水
	MoneyRecordsCreateFailedErr  = 2000 //流水创建失败
	CoachBalanceNotEnoughErr     = 2001 //教练余额不足
	CoachDepositNotEnoughErr     = 2002 //教练保证金不足
	ClubBalanceNotEnoughErr      = 2004 //俱乐部余额不足
	ClubDepositNotEnoughErr      = 2003 //俱乐部保证金不足
	MoneyOperateOrderNotExistErr = 2005 //保证金充值订单不存在

	//推荐码
	ReferralCodeNotExistErr = 2100 //推荐码不存在

	//打包课
	OrderTransferErr = 2200 //订单转包课失败

	//第三方错误码错
	WeChatLoginErr   = 10001 //微信登录失败
	WeChatDecryptErr = 10002 //微信解密失败密失败

	WechatTransferBillErr = 10003
)

type Err struct {
	Code int
	Msg  string
}

func (e *Err) Error() string {
	return e.Msg
}
func (e *Err) GetCode() int {
	return e.Code
}
func (e *Err) GetMsg() string {
	return e.Msg
}

func NewErr(code int, msg string) *Err {
	return &Err{
		Code: code,
		Msg:  msg,
	}
}
