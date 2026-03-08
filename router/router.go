package router

import (
	"skis-admin-backend/controller"
	"skis-admin-backend/middlewares"

	"github.com/gin-gonic/gin"
)

func FeedsRouter(Router *gin.RouterGroup) {
	RouteRouter := Router.Group("feeds")
	{
		//动态接口
		RouteRouter.POST("", middlewares.JWTAuth(), controller.CreateFeed) //创建动态
		RouteRouter.GET("", controller.QueryFeedsList)                     //查询动态列表
		RouteRouter.GET("/:id", controller.QueryFeedInfo)                  //查询动态信息

		RouteRouter.PUT("/:id", middlewares.JWTAuth(), controller.UpdateFeed)    //更新动态信息
		RouteRouter.DELETE("/:id", middlewares.JWTAuth(), controller.DeleteFeed) //删除动态

		RouteRouter.GET("/:id/comments", middlewares.JWTAuth(), controller.QueryFeedsCommentsList)             //查询动态圈评论列表
		RouteRouter.GET("/:id/comments/:comment_id", middlewares.JWTAuth(), controller.QueryFeedsCommentInfo)  //查询动态圈评论信息
		RouteRouter.POST("/:id/comments", middlewares.JWTAuth(), controller.CreateFeedsComments)               //评论动态圈
		RouteRouter.PUT("/:id/comments/:comment_id", middlewares.JWTAuth(), controller.UpdateFeedsComments)    //更新动态圈评论
		RouteRouter.DELETE("/:id/comments/:comment_id", middlewares.JWTAuth(), controller.DeleteFeedsComments) //删除动态圈评论
	}
}

func ClubsRouter(Router *gin.RouterGroup) {
	RouteRouter := Router.Group("clubs")
	{
		//俱乐部接口
		RouteRouter.GET("/info", middlewares.JWTAuth(), controller.QueryClubInfo)                      //查询俱乐部信息
		RouteRouter.POST("/login", controller.ClubsUserLogin)                                          //俱乐部用户登录
		RouteRouter.GET("", middlewares.JWTAuth(), controller.QueryClubsList)                          //查询俱乐部列表
		RouteRouter.PUT("", middlewares.JWTAuth(), controller.UpdateClubsInfo)                         //更新俱乐部
		RouteRouter.GET("/goods", middlewares.JWTAuth(), controller.QueryClubGoods)                    //查询俱乐部商品
		RouteRouter.GET("/:club_id", middlewares.JWTAuth(), controller.QueryClubInfo)                  //查询俱乐部信息
		RouteRouter.GET("/:club_id/goods", middlewares.JWTAuth(), controller.QueryOneClubGoods)        //查询指定俱乐部商品
		RouteRouter.POST("/apply", middlewares.JWTAuth(), controller.ApplyClub)                        //申请成为俱乐部
		RouteRouter.GET("/apply", middlewares.JWTAuth(), controller.QueryApplyClubInfo)                //查询俱乐部申请信息
		RouteRouter.GET("/order_course", middlewares.JWTAuth(), controller.QueryClubOrderCourses)      //查询俱乐部课程
		RouteRouter.GET("/feeds", middlewares.JWTAuth(), controller.QueryClubFeedsList)                //查询俱乐部动态列表
		RouteRouter.GET("/courses", middlewares.JWTAuth(), controller.QueryClubsCourses)               //查询俱乐部课程
		RouteRouter.GET("/match/coaches", middlewares.JWTAuth(), controller.QueryClubMatchCoachesList) //查询俱乐部匹配教练列表

		RouteRouter.GET("/:club_id/comments", middlewares.JWTAuth(), controller.QueryClubComments)   //获取俱乐部评论
		RouteRouter.GET("/:club_id/ski_resorts_list", controller.ClubGetSkiResorts)                  //获取俱乐部雪场列表
		RouteRouter.GET("/:club_id/ski_resorts_date_list", controller.ClubGetSkiResortDate)          //获取俱乐部雪场日期列表
		RouteRouter.GET("/:club_id/ski_resorts_time_list", controller.ClubGetSkiResortTime)          //获取俱乐部雪场时间列表
		RouteRouter.GET("/teach_date", middlewares.JWTAuth(), controller.ClubSkiResortTeachDateList) //查询俱乐部雪场教学日期

		RouteRouter.GET("/coach/list", middlewares.JWTAuth(), controller.ClubsCoachList)                                        //俱乐部的教练列表
		RouteRouter.POST("/:id/coach", middlewares.JWTAuth(), controller.ClubCheckCoach)                                        //俱乐部审核教练请求
		RouteRouter.POST("/change_teach_time/:order_course_id", middlewares.JWTAuth(), controller.ClubChangeTeachTime)          //俱乐部修改学员预约的上课时间
		RouteRouter.POST("/appointment_coach_course/:order_course_id", middlewares.JWTAuth(), controller.ClubAppointmentCourse) //安排教练上课
		RouteRouter.POST("/replace_coach_course/:order_course_id", middlewares.JWTAuth(), controller.ClubReplaceCoachCourse)    //更换教练上课

		//RouteRouter.POST("/coach/quit", middlewares.JWTAuth(), controller.CoachQuitClub) //教练退出俱乐部

		//RouteRouter.POST("", controller.CreateClub)       //创建俱乐部
		//RouteRouter.PUT("/:club_id", controller.UpdateClub) //更新俱乐部信息
	}
}

func OrdersRouter(Router *gin.RouterGroup) {
	RouteRouter := Router.Group("orders")
	{
		RouteRouter.POST("", middlewares.JWTAuth(), controller.CreateOrder)                           //创建订单
		RouteRouter.GET("", middlewares.JWTAuth(), controller.QueryOrdersList)                        //查询订单列表
		RouteRouter.GET("/:order_id", middlewares.JWTAuth(), controller.QueryOrderInfo)               //查询订单信息
		RouteRouter.POST("/pay_callback", controller.OrderPayCallback)                                //订单支付回调
		RouteRouter.POST("/whirly_pay_callback", controller.WhirlyOrderPayCallback)                   //模拟订单支付回调
		RouteRouter.GET("/:order_id/refund", middlewares.JWTAuth(), controller.QueryOrdersRefundInfo) //订单退款
		RouteRouter.POST("/:order_id/refund", middlewares.JWTAuth(), controller.OrderRefund)          //订单退款
		RouteRouter.POST("/refund_callback", controller.OrderRefundCallback)                          //订单退款回调
		RouteRouter.POST("/refund_callback_test", controller.OrderRefundCallbackTest)                 //订单退款回调测试
		RouteRouter.POST("/test/whirly", middlewares.JWTAuth(), controller.TestWhirly)                //流水添加测试

		RouteRouter.GET("/order_course", middlewares.JWTAuth(), controller.QueryOrderCoursesList) //查询订单课程列表

		//RouteRouter.PUT("/:id", controller.UpdateOrder)    //更新订单信息
	}
}

func OrdersCoursesRouter(Router *gin.RouterGroup) {
	RouteRouter := Router.Group("orders_courses")
	{
		RouteRouter.GET("/:order_course_id", middlewares.JWTAuth(), controller.QueryOrdersCourseInfo)                                //查询订单课程记录列表
		RouteRouter.GET("/:order_course_id/records", middlewares.JWTAuth(), controller.QueryOrdersCoursesRecords)                    //查询订单课程记录列表
		RouteRouter.POST("/:order_course_id/records", middlewares.JWTAuth(), controller.CreateOrdersCoursesRecord)                   //创建订单课程记录
		RouteRouter.GET("/:order_course_id/comment", middlewares.JWTAuth(), controller.QueryOrdersCoursesComment)                    //查询订单课程评论
		RouteRouter.POST("/:order_course_id/comment", middlewares.JWTAuth(), controller.CreateOrdersCoursesComment)                  //创建订单课程评论
		RouteRouter.POST("/:order_course_id/comment/:comment_id/reply", middlewares.JWTAuth(), controller.ReplyOrdersCoursesComment) //回复订单课程评论
	}
}

func OrdersCoursesRecordsRouter(Router *gin.RouterGroup) {
	RouteRouter := Router.Group("orders_courses_records")
	{
		RouteRouter.DELETE("/:id", middlewares.JWTAuth(), controller.DeleteOrdersCoursesRecord) //删除用户订单课程记录
	}
}

func GoodsRouter(Router *gin.RouterGroup) {
	RouteRouter := Router.Group("goods")
	{
		RouteRouter.GET("", controller.QueryGoodsList) //查询商品列表

		RouteRouter.POST("/create", middlewares.JWTAuth(), controller.CreateGoods)          //创建商品
		RouteRouter.POST("/edit", middlewares.JWTAuth(), controller.EditGoods)              //编辑商品
		RouteRouter.POST("/before_edit", middlewares.JWTAuth(), controller.BeforeEditGoods) //编辑商品
		RouteRouter.POST("/take_down", middlewares.JWTAuth(), controller.TakeDownGoods)     //下架商品
		RouteRouter.POST("/put_up", middlewares.JWTAuth(), controller.PutUpGoods)           //上架商品

		RouteRouter.POST("/delete", middlewares.JWTAuth(), controller.DeleteGoods)

		RouteRouter.GET("/:good_id", middlewares.JWTAuth(), controller.QueryGoodInfo) //查询商品信息
		//RouteRouter.GET("/:good_id/before_buy", middlewares.JWTAuth(), controller.QueryGoodBeforeBuy) //查询商品信息

	}
}

func CoachesRouter(Router *gin.RouterGroup) {
	RouteRouter := Router.Group("coaches")
	{
		//教练接口
		RouteRouter.GET("", controller.QueryCoachesList)
		RouteRouter.GET("/apply", middlewares.JWTAuth(), controller.QueryApplyInfo)                       //查询教练申请信息
		RouteRouter.POST("/apply", middlewares.JWTAuth(), controller.ApplyCoach)                          //申请成为教练
		RouteRouter.GET("/info", middlewares.JWTAuth(), controller.QueryLoginCoachInfo)                   //查询教练列表
		RouteRouter.GET("/courses", middlewares.JWTAuth(), controller.QueryCoachCourses)                  //查询教练课程
		RouteRouter.GET("/order_course", middlewares.JWTAuth(), controller.QueryCoachOrderCourses)        //查询教练课程
		RouteRouter.GET("/goods", middlewares.JWTAuth(), controller.QueryCoachGoods)                      //查询教练课程
		RouteRouter.GET("/feeds", middlewares.JWTAuth(), controller.QueryCoachFeedsList)                  //查询教练动态列表
		RouteRouter.GET("/referral_records", middlewares.JWTAuth(), controller.QueryCoachReferralRecords) //查询教练推荐记录
		RouteRouter.GET("/:coach_id/goods", controller.QueryOneCoachGoods)                                //查询教练信息
		RouteRouter.GET("/:coach_id", controller.QueryCoachInfo)                                          //查询教练信息
		RouteRouter.GET("/match", middlewares.JWTAuth(), controller.QueryMatchCoachesList)                //查询匹配教练列表
		RouteRouter.POST("/:coach_id", middlewares.JWTAuth(), controller.EditCoachInfo)                   //编辑教练信息
		RouteRouter.GET("/:coach_id/comments", middlewares.JWTAuth(), controller.QueryCoachComments)      //获取教练评论

		RouteRouter.POST("/ski_resorts/edit", middlewares.JWTAuth(), controller.CoachEditSkiResorts) //教练修改雪场
		RouteRouter.GET("/:coach_id/ski_resorts_list", controller.CoachGetSkiResorts)                //获取教练雪场列表
		RouteRouter.GET("/:coach_id/ski_resorts_date_list", controller.CoachGetSkiResortDate)        //获取教练雪场日期列表
		RouteRouter.GET("/:coach_id/ski_resorts_time_list", controller.CoachGetSkiResortTime)        //获取教练雪场时间列表

		RouteRouter.GET("/tag", middlewares.JWTAuth(), controller.CoachGetAllTags)               //获取教练所有技能
		RouteRouter.POST("/tag/add_review", middlewares.JWTAuth(), controller.CoachAddTagReview) //教练申请技能
		RouteRouter.POST("/tag/get_review", middlewares.JWTAuth(), controller.CoachGetTagReview) //获取教练申请技能记录
		RouteRouter.POST("/tag/remove", middlewares.JWTAuth(), controller.CoachRemoveTag)        //教练取消技能

		RouteRouter.POST("/certificate/add_review", middlewares.JWTAuth(), controller.CoachAddCertificates) //教练申请证书
		RouteRouter.POST("/certificate/get_review", middlewares.JWTAuth(), controller.CoachGetCertificates) //获取教练证书审核记录
		RouteRouter.GET("/certificate", middlewares.JWTAuth(), controller.CoachGetAllCertificates)          //教练所有证书

		RouteRouter.POST("/club/join", middlewares.JWTAuth(), controller.CoachJoinClubs) //教练申请加入俱乐部
		RouteRouter.POST("/club/quit", middlewares.JWTAuth(), controller.CoachQuitClubs) //教练退出俱乐部
		RouteRouter.POST("/club/list", middlewares.JWTAuth(), controller.CoachClubsList) //获取教练俱乐部列表

		RouteRouter.POST("/confirm_order_course/:order_course_id", middlewares.JWTAuth(), controller.CoachConfirmOrderCourses)                   //教练确认预约课程
		RouteRouter.GET("/before_change_order_course_time/:order_course_id", middlewares.JWTAuth(), controller.BeforeCoachChangeOrderCourseTime) //教练修改预约课程时间之前
		RouteRouter.POST("/change_order_course_time/:order_course_id", middlewares.JWTAuth(), controller.CoachChangeOrderCourseTime)             //教练修改预约课程时间
		RouteRouter.POST("/before_cancel_order_course/:order_course_id", middlewares.JWTAuth(), controller.BeforeCoachCancelOrderCourses)        //教练取消预约课程之前
		RouteRouter.POST("/cancel_order_course/:order_course_id", middlewares.JWTAuth(), controller.CoachCancelOrderCourses)                     //教练取消预约课程
		RouteRouter.POST("/transfer_order_course/:order_course_id", middlewares.JWTAuth(), controller.CoachTransferOrderCourses)                 //教练申请转单
		RouteRouter.POST("/cancel_transfer_order_course/:order_course_id", middlewares.JWTAuth(), controller.CoachCancelTransferOrderCourses)    //取消转单课程
		RouteRouter.POST("/transfer_order_to_coach", middlewares.JWTAuth(), controller.CoachTransferOrderToCoach)                                //转单给教练
		RouteRouter.POST("/review_order_from_coach", middlewares.JWTAuth(), controller.CoachReviewOrderFromCoach)                                //教练审核其他教练转过来的课程

		RouteRouter.POST("/review_order_from_club", middlewares.JWTAuth(), controller.CoachReviewOrderFromClub)                             //教练审核俱乐部转过来的课程
		RouteRouter.POST("/apply_transfer_order/:order_course_id", middlewares.JWTAuth(), controller.CoachApplyTransferOrders)              //教练申请转单（待俱乐部审核）
		RouteRouter.POST("/cancel_apply_transfer_order/:order_course_id", middlewares.JWTAuth(), controller.CoachCancelApplyTransferOrders) //教练取消申请转单（待俱乐部审核）
		RouteRouter.POST("/review_replace_from_club", middlewares.JWTAuth(), controller.CoachReviewReplaceFromClub)                         //教练审核俱乐部更换教练的课程

		RouteRouter.POST("/verify_order_course/:check_code", middlewares.JWTAuth(), controller.CoachVerifyCourses) //教练核销课程

	}
}
func TagsRouter(Router *gin.RouterGroup) {
	RouteRouter := Router.Group("tags")
	{
		//标签接口s
		RouteRouter.GET("", controller.QueryTagsList)    //查询标签列表
		RouteRouter.GET("/:id", controller.QueryTagInfo) //查询标签信息
	}
}

func SkiResortRouter(Router *gin.RouterGroup) {
	RouteRouter := Router.Group("ski_resorts")
	{
		//雪场接口
		RouteRouter.GET("", controller.QuerySkiResortsList)                                           //查询雪场列表
		RouteRouter.GET("/teach_time", middlewares.JWTAuth(), controller.QuerySkiResortTeachTimeList) //查询雪场教学时间
		RouteRouter.GET("/teach_date", middlewares.JWTAuth(), controller.QuerySkiResortTeachDateList) //查询雪场教学日期
		RouteRouter.POST("/teach_time", middlewares.JWTAuth(), controller.CreateSkiResortTeachTime)   //创建雪场教学时间
		RouteRouter.POST("/teach_state", middlewares.JWTAuth(), controller.UpdateSkiResortTeachState) //更新雪场教学状态
		RouteRouter.DELETE("/teach_time", middlewares.JWTAuth(), controller.DeleteSkiResortTeachTime) //删除雪场教学时间
		RouteRouter.GET("/schedule_event", middlewares.JWTAuth(), controller.ScheduleEvent)           //日程事件
		RouteRouter.POST("/quit_club", middlewares.JWTAuth(), controller.QuitClub)                    //退出雪场
		RouteRouter.GET("/:id", controller.QuerySkiResortInfo)                                        //查询雪场信息
		//RouteRouter.POST("", controller.CreateSkiResort)       //创建雪场
		//RouteRouter.PUT("/:id", controller.UpdateSkiResort)    //更新雪场信息
		//RouteRouter.DELETE("/:id", controller.DeleteSkiResort) //删除雪场
	}
}

func MoneyRouter(Router *gin.RouterGroup) {
	RouteRouter := Router.Group("money")
	{
		RouteRouter.POST("/verify/course/:order_course_id", middlewares.JWTAuth(), controller.MoneyTest) //用户核销课程
		//资金接口
		RouteRouter.GET("", middlewares.JWTAuth(), controller.QueryMoneyList)           //查询资金列表
		RouteRouter.GET("/:money_id", middlewares.JWTAuth(), controller.QueryMoneyInfo) //查询资金信息

		//资金提现接口
		RouteRouter.POST("/operate/withdraw", middlewares.JWTAuth(), controller.MoneyOperateWithdraw) //资金操作提现
		RouteRouter.POST("/operate/recharge", middlewares.JWTAuth(), controller.MoneyOperateRecharge) //资金操作充值

		//资金提现回调
		RouteRouter.POST("/withdraw/pay_callback", controller.MoneyOperateWithdraw) //资金操作提现
		//资金充值回调
		RouteRouter.POST("/deposit/pay_callback", controller.DepositPayCallback) //资金操作充值保证金
		RouteRouter.POST("/deposit/test_pay", controller.DepositTestPay)         //资金操作充值
		RouteRouter.POST("/transfer/bills", controller.TransferBills)            //资金操作充值
	}

}

func UsersRouter(Router *gin.RouterGroup) {
	RouteRouter := Router.Group("users")
	{
		//用户接口
		RouteRouter.POST("/login", controller.Login)                                                                     //用户登录接口
		RouteRouter.GET("", middlewares.JWTAuth(), controller.QueryUserInfo)                                             //用户登录接口
		RouteRouter.PUT("", middlewares.JWTAuth(), controller.UpdateUserInfo)                                            //更新用户信息
		RouteRouter.POST("/active", middlewares.JWTAuth(), controller.UserActive)                                        //用户活跃接口
		RouteRouter.GET("/points_records", middlewares.JWTAuth(), controller.QueryPointsRecordsList)                     //查询用户积分记录
		RouteRouter.GET("/orders_courses_records", middlewares.JWTAuth(), controller.QueryUserOrdersCoursesRecords)      //查询订单课程记录列表
		RouteRouter.GET("/:uid/orders_courses_records", middlewares.JWTAuth(), controller.QueryUserOrdersCoursesRecords) //查询用户订单课程记录列表
		RouteRouter.GET("/:uid", middlewares.JWTAuth(), controller.QueryUserInfo)                                        //查询用户信息
		RouteRouter.PUT("/phone", middlewares.JWTAuth(), controller.UpdateUserPhone)                                     //更新用户手机号

		//预约课程接口
		RouteRouter.POST("/appointment_course", middlewares.JWTAuth(), controller.AppointmentCourse)                           //用户预约课程
		RouteRouter.POST("/before_cancel_appointment_course", middlewares.JWTAuth(), controller.BeforeCancelAppointmentCourse) //用户取消预约课程之前调用的接口
		RouteRouter.POST("/cancel_appointment_course", middlewares.JWTAuth(), controller.CancelAppointmentCourse)              //用户取消预约课程

		RouteRouter.POST("/review_teach_time", middlewares.JWTAuth(), controller.ReviewTeachTime)                    //用户审核修改上课时间
		RouteRouter.POST("/review_coach_transfer_order", middlewares.JWTAuth(), controller.ReviewCoachTransferOrder) //用户审核教练转单

		//用户核销课程接口
		RouteRouter.POST("/verify_course/:order_course_id", middlewares.JWTAuth(), controller.VerifyCourse) //用户核销课程

	}
}

// 下面暂时不用的代码

func CoursesRouter(Router *gin.RouterGroup) {
	RouteRouter := Router.Group("courses")
	{
		//课程接口
		RouteRouter.GET("", controller.QueryCoursesList)           //查询课程列表
		RouteRouter.GET("/:course_id", controller.QueryCourseInfo) //查询课程信息
	}
}

func CoachLevelRouter(Router *gin.RouterGroup) {
	RouteRouter := Router.Group("coaches_level")
	{
		//教练等级接口
		RouteRouter.GET("", controller.QueryCoachesLevelsList) //查询教练等级列表
	}
}

func PointsRecordsRouter(Router *gin.RouterGroup) {
	RouteRouter := Router.Group("points_records")
	{
		//积分记录接口
		RouteRouter.GET("", middlewares.JWTAuth(), controller.QueryPointsRecordsList)          //查询积分记录列表
		RouteRouter.GET("/:point_id", middlewares.JWTAuth(), controller.QueryPointsRecordInfo) //查询积分记录信息
	}
}

func CertificateConfigsRouter(Router *gin.RouterGroup) {
	RouteRouter := Router.Group("certificate_configs")
	{
		//证书标签接口
		RouteRouter.GET("", controller.QueryCertificateConfigsList)    //查询证书标签列表
		RouteRouter.GET("/:id", controller.QueryCertificateConfigInfo) //查询证书标签信息
	}
}

func OssRouter(Router *gin.RouterGroup) {
	RouteRouter := Router.Group("oss")
	{
		//oss接口
		RouteRouter.GET("/sign", middlewares.JWTAuth(), controller.GetOssSign) //获取oss签名
	}
}

func AdminRouter(Router *gin.RouterGroup) { //管理台的接口
	RouteRouter := Router.Group("admin")
	{
		//管理员接口
		RouteRouter.GET("/orders/:order_id/refund/limit", middlewares.LocalAuth(), controller.AdminQueryOrderRefundLimit) //查询订单退款限制
		RouteRouter.POST("/orders/:order_id/refund", middlewares.LocalAuth(), controller.AdminOrderRefund)                //退款
		RouteRouter.GET("/orders/:order_id/match_coaches", middlewares.LocalAuth(), controller.AdminMatchCoachesList)
		RouteRouter.POST("/orders/:order_id/transfer", middlewares.LocalAuth(), controller.AdminTransferOrderToCoach) //转单给教练
	}
}
