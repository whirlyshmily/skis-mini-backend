package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"skis-admin-backend/enum"
	"skis-admin-backend/global"
	"skis-admin-backend/middlewares"
	"skis-admin-backend/response"
	"time"
)

func CreateToken(c *gin.Context, userId, openId, unionId string, userType int) string {
	//生成token信息
	j := middlewares.NewJWT()
	claims := middlewares.CustomClaims{
		Uid:      userId,
		OpenId:   openId,
		UnionId:  unionId,
		UserType: userType,
		StandardClaims: jwt.StandardClaims{
			NotBefore: time.Now().Unix(),
			// TODO 设置token过期时间
			ExpiresAt: time.Now().Unix() + 60*60*24*7, //token -->7天过期
			Issuer:    "test",
		},
	}
	//生成token
	token, err := j.CreateToken(claims)
	if err != nil {
		response.Err(c, err)
		return ""
	}
	return token
}

func GetNowFormatTodayTime() string {
	now := time.Now()
	format := "2006-01-02"
	formattedTime := now.Format(format)
	return formattedTime
}

// 请求参数重校验claims
func GetClaimsFromContext(c *gin.Context) (*middlewares.CustomClaims, bool) {
	claims, exists := c.Get("claims")
	if !exists {
		// 如果不存在，说明中间件没有设置claims
		global.Lg.Error("未登录")
		response.Err(c, enum.NewErr(enum.TokenInvalidErr, "token无效"))
		return nil, false
	}

	// 进行类型断言，确保 claims 的类型是 jwt.Claims
	jwtClaims, ok := claims.(*middlewares.CustomClaims)
	if !ok {
		global.Lg.Error("未登录")
		response.Err(c, enum.NewErr(enum.TokenInvalidErr, "token无效"))
		return nil, false
	}

	return jwtClaims, true
}
