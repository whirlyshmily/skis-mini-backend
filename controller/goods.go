package controller

import (
	"skis-admin-backend/dao"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/response"

	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"go.uber.org/zap"
)

func QueryGoodsList(c *gin.Context) {
	var req forms.QueryGoodsListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	list, err := dao.NewGoodsDao(c, global.DB).ListByUserId(c, req)
	if err != nil {
		global.Lg.Error("查询动态列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, &forms.QueryGoodsListResponse{
		List: list,
	})
	return
}

func CreateGoods(c *gin.Context) {
	var req forms.SaveGoodsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	goods, err := dao.NewGoodsDao(c, global.DB).CreateGoods(c, req)
	if err != nil {
		global.Lg.Error("创建商品失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, goods)
	return
}

func EditGoods(c *gin.Context) {
	var req forms.SaveGoodsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	goods, err := dao.NewGoodsDao(c, global.DB).EditGoods(c, req)
	if err != nil {
		global.Lg.Error("编辑商品失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, goods)
	return
}

func BeforeEditGoods(c *gin.Context) {
	var req forms.BeforeEditGoodsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	goods, err := dao.NewGoodsDao(c, global.DB).BeforeEditGoods(c, req.GoodID)
	if err != nil {
		global.Lg.Error("编辑商品失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, goods)
	return
}

func TakeDownGoods(c *gin.Context) {
	var req forms.TakeDownGoodsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	err := dao.NewGoodsDao(c, global.DB).TakeDownGoods(c, req)
	if err != nil {
		global.Lg.Error("下架商品失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, "商品下架成功")
	return
}

func PutUpGoods(c *gin.Context) {
	var req forms.PutUpGoodsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	err := dao.NewGoodsDao(c, global.DB).PutUpGoods(c, req)
	if err != nil {
		global.Lg.Error("上架商品失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, "商品上架成功")
	return
}

func QueryGoodInfo(c *gin.Context) {
	goodId := c.Param("good_id")
	good, err := dao.NewGoodsDao(c, global.DB).QueryGoodInfo(goodId)
	if err != nil {
		global.Lg.Error("查询商品详情失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	uid := c.GetString("uid")
	good.PayData, err = dao.NewGoodsDao(c, global.DB).QueryGoodBeforeBuy(uid, good)
	if err != nil {
		global.Lg.Error("QueryGoodBeforeBuy 查询商品详情失败", zap.Error(err))
	}
	response.Success(c, good)
	return
}

//	func QueryGoodBeforeBuy(c *gin.Context) {
//		goodId := c.Param("good_id")
//		uid := c.GetString("uid")
//		good, err := dao.NewGoodsDao(c, global.DB).QueryGoodBeforeBuy(uid, goodId)
//		if err != nil {
//			global.Lg.Error("查询商品详情失败", zap.Error(err))
//			response.Err(c, err)
//			return
//		}
//		response.Success(c, good)
//		return
//	}
func DeleteGoods(c *gin.Context) {
	var req forms.DeleteGoodsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}
	err := dao.NewGoodsDao(c, global.DB).DeleteGoods(c, req)
	if err != nil {
		global.Lg.Error("删除商品失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, "商品删除成功")
	return
}

func QueryCoachGoods(c *gin.Context) {
	coachId := c.GetString("coach_id")
	if coachId == "" {
		response.Err(c, enum.NewErr(enum.TokenInvalidErr, "用户ID不存在"))
		return
	}

	var req forms.QueryGoodsListRequest
	req.UserId = coachId
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Err(c, err)
		return
	}

	goods, err := dao.NewGoodsDao(c, global.DB).ListByUserId(c, req)
	if err != nil {
		global.Lg.Error("查询课程列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, goods)
	return
}
func QueryOneCoachGoods(c *gin.Context) {
	coachId := c.Param("coach_id")
	if coachId == "" {
		response.Err(c, enum.NewErr(enum.TokenInvalidErr, "用户ID不存在"))
		return
	}

	var req forms.QueryGoodsListRequest
	req.UserId = coachId
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Err(c, err)
		return
	}
	req.OnShelf = core.Int32(1)

	goods, err := dao.NewGoodsDao(c, global.DB).ListByUserId(c, req)
	if err != nil {
		global.Lg.Error("查询课程列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, goods)
	return
}

func QueryClubGoods(c *gin.Context) {
	clubId := c.GetString("club_id")
	if clubId == "" {
		response.Err(c, enum.NewErr(enum.TokenInvalidErr, "用户ID不存在"))
		return
	}

	var req forms.QueryGoodsListRequest
	req.UserId = clubId
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Err(c, err)
		return
	}

	goods, err := dao.NewGoodsDao(c, global.DB).ListByUserId(c, req)
	if err != nil {
		global.Lg.Error("查询课程列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, goods)
	return
}

func QueryOneClubGoods(c *gin.Context) {
	var req forms.QueryGoodsListRequest
	req.UserId = c.Param("club_id")
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Err(c, err)
		return
	}

	req.OnShelf = core.Int32(1) //只返回上架商品
	goods, err := dao.NewGoodsDao(c, global.DB).ListByUserId(c, req)
	if err != nil {
		global.Lg.Error("查询课程列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, goods)
	return
}
