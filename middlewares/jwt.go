package middlewares

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"skis-admin-backend/dao"
	"skis-admin-backend/enum"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
	"skis-admin-backend/response"
	"strings"
	"time"
)

type JWT struct {
	SigningKey []byte
}

type CustomClaims struct {
	Uid      string
	OpenId   string
	UnionId  string
	UserType int
	jwt.StandardClaims
}

var (
	TokenInvalid = errors.New("Couldn't handle this token:")
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取 Authorization 头部
		authorization := c.GetHeader("Authorization")
		if authorization == "" {
			response.Err(c, enum.NewErr(enum.TokenInvalidErr, "token 不能为空"))
			c.Abort()
			return
		}
		// parseToken 解析token包含的信息
		token := ExtractTokenFromHeader(authorization)
		j := NewJWT()
		claims, err := j.ParseToken(token)
		jwt.TimeFunc = time.Now
		if err != nil {
			response.Err(c, err)
			c.Abort()
			return
		}
		// gin的上下文记录claims和userId的值
		c.Set("claims", claims)
		c.Set("uid", claims.Uid)
		c.Set("open_id", claims.OpenId)
		c.Set("union_id", claims.UnionId)
		c.Set("user_type", claims.UserType)
		c.Set("user_id", claims.Uid)
		c.Set("app_id", global.Config.UserMiniProgram.AppId)
		//根据userType 查询是教练还是用户
		if claims.UserType == model.UserTypeUser {
			coach, err := dao.CoachInfoByUserId(claims.Uid)
			if err != nil && err != gorm.ErrRecordNotFound {
				global.Lg.Error("查询教练失败", zap.Error(err))
				response.Err(c, err)
				c.Abort()
				return
			}
			if coach != nil {
				c.Set("coach_id", coach.CoachId)
				c.Set("user_type", model.UserTypeCoach)
				c.Set("user_id", coach.CoachId)
			}

		} else if claims.UserType == model.UserTypeClub {
			c.Set("app_id", global.Config.ClubMiniProgram.AppId)
			club, err := dao.QueryClubInfoByUid(claims.Uid)
			if err != nil && err != gorm.ErrRecordNotFound {
				global.Lg.Error("查询俱乐部失败", zap.Error(err))
				response.Err(c, err)
				c.Abort()
				return
			}
			if club != nil {
				c.Set("club_id", club.ClubId)
				c.Set("user_type", model.UserTypeClub)
				c.Set("user_id", club.ClubId)
			}
		}
		c.Next()
	}
}

// 辅助函数：从 Authorization 头部中提取 Token
func ExtractTokenFromHeader(authHeader string) string {
	// Token 应该以 "Bearer " 前缀开始，因此我们可以简单地删除前缀以获取 Token
	return strings.TrimPrefix(authHeader, "Bearer ")
}

func NewJWT() *JWT {
	return &JWT{
		[]byte(global.Config.JWT.SigningKey),
	}
}

// 创建一个token
func (j *JWT) CreateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// 解析 token
func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, enum.NewErr(enum.TokenInvalidErr, "token无效")
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, enum.NewErr(enum.TokenInvalidErr, "token无效")
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, enum.NewErr(enum.TokenInvalidErr, "token无效")
			} else {
				return nil, enum.NewErr(enum.TokenInvalidErr, "token无效")
			}
		}
	}
	if token != nil {
		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			return claims, nil
		}
		return nil, enum.NewErr(enum.TokenInvalidErr, "token无效")

	} else {
		return nil, enum.NewErr(enum.TokenInvalidErr, "token无效")
	}
}

// 更新token
func (j *JWT) RefreshToken(tokenString string) (string, error) {
	jwt.TimeFunc = func() time.Time {
		return time.Unix(0, 0)
	}
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		jwt.TimeFunc = time.Now
		claims.StandardClaims.ExpiresAt = time.Now().Add(1 * time.Hour).Unix()
		return j.CreateToken(*claims)
	}
	return "", TokenInvalid
}
