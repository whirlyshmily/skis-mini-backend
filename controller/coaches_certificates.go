package controller

import (
	"errors"
	"github.com/gin-gonic/gin"
	"skis-admin-backend/dao"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/response"
	"strconv"
)

func CoachAddCertificates(c *gin.Context) {
	var req forms.CoachAddCertificatesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, err)
		return
	}

	err := dao.NewCoachesCertificatesDao(c, global.DB).CoachAddCertificates(c, req)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, "证书申请添加成功")
	return
}

func CoachGetAllCertificates(c *gin.Context) {
	coachId := c.GetString("coach_id")
	if coachId == "" {
		response.Err(c, errors.New("请先申请成为教练"))
	}
	verified, err := strconv.Atoi(c.Query("verified"))
	if err != nil {
		response.Err(c, err)
		return
	}
	items, err := dao.NewCoachesCertificatesDao(c, global.DB).CoachGetAllCertificates(coachId, verified)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, items)
	return
}

func CoachGetCertificates(c *gin.Context) {
	var req forms.CoachGetCertificatesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, err)
		return
	}

	items, err := dao.NewCoachesCertificatesDao(c, global.DB).CoachGetCertificates(c, req)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.Success(c, items)
	return
}
