package controller

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"skis-admin-backend/dao"
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/response"
	"strconv"
)

func QueryCertificateConfigsList(c *gin.Context) {
	var req forms.QueryCertificateConfigsListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("参数校验失败", zap.Error(err))
		response.Err(c, enum.NewErr(enum.ParamErr, "参数校验失败"))
		return
	}

	tags, err := dao.QueryCertificateConfigsList(req)
	if err != nil {
		global.Lg.Error("查询证书配置列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, forms.QueryCertificateConfigsListResponse{
		List: tags,
	})
	return
}

func QueryCertificateConfigInfo(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, enum.NewErr(enum.ParamErr, "参数错误"))
		return
	}
	tag, err := dao.QueryCertificateConfigInfo(id)
	if err != nil {
		global.Lg.Error("查询证书配置信息失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, tag)
	return
}
