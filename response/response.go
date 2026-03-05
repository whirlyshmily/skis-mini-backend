package response

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	"net/http"
	"skis-admin-backend/enum"
)

type BaseResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// 返回成功
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, BaseResponse{
		Code: enum.Success,
		Msg:  "success",
		Data: data,
	})
}

// 返回失败
func Err(c *gin.Context, err error) {
	code := enum.GeneralErr
	msg := err.Error()

	if errors.Is(err, gorm.ErrRecordNotFound) {
		code = enum.DataNotExist
		msg = gorm.ErrRecordNotFound.Error()
	} else if r, ok := err.(*enum.Err); ok { //项目直接返回estorError的时候，这里兼容处理
		code = r.GetCode()
		msg = r.GetMsg()
	} else if validatorErr, ok := err.(validator.ValidationErrors); ok {
		code = enum.ParamErr
		msg = validatorErr.Error()
	}
	c.JSON(http.StatusOK, BaseResponse{
		Code: code,
		Msg:  msg,
	})
}
