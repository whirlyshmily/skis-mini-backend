package controller

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"skis-admin-backend/dao"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/response"
	"strconv"
)

func QueryTagsList(c *gin.Context) {
	var req forms.QueryTagsListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	total, list, err := dao.QueryTagsList(&req)
	if err != nil {
		global.Lg.Error("查询标签列表失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, &forms.QueryTagsListResponse{
		List:  list,
		Total: total,
	})
	return
}

func CreateTag(c *gin.Context) {
	var req forms.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	tag, err := dao.CreateTag(req)
	if err != nil {
		global.Lg.Error("创建标签失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, tag)
	return
}

func UpdateTag(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	var req forms.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	if err = dao.UpdateTag(id, &req); err != nil {
		global.Lg.Error("更新标签失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, nil)
	return
}

func QueryTagInfo(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	tag, err := dao.QueryTagInfo(id)
	if err != nil {
		global.Lg.Error("查询标签失败", zap.Error(err))
		response.Err(c, err)
		return
	}
	response.Success(c, tag)
	return
}

func DeleteTag(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		global.Lg.Error("参数错误", zap.Error(err))
		response.Err(c, err)
		return
	}

	if err = dao.DeleteTag(id); err != nil {
		global.Lg.Error("删除标签失败", zap.Error(err))
		response.Err(c, err)
		return
	}

	response.Success(c, nil)
	return
}
