package dao

import (
	"skis-admin-backend/enum"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func QueryFeedsCommentsList(c *gin.Context, feedId int64, req *forms.QueryFeedsCommentsListRequest) ([]*model.FeedsComments, error) {
	var comments []*model.FeedsComments
	db := global.DB.Model(&model.FeedsComments{}).Where("feed_id = ? and state = 0 and on_shelf = 1", feedId).Order("id desc")
	if req.LastId > 0 {
		db = db.Where("id < ?", req.LastId)
	}

	if err := db.Limit(req.PageSize).Find(&comments).Error; err != nil {
		global.Lg.Error("QueryFeedsComments error", zap.Error(err))
		return nil, err
	}

	for _, comment := range comments {
		checkRight(c, comment)
	}

	return comments, nil
}

func checkRight(c *gin.Context, comment *model.FeedsComments) bool {
	if c.GetInt("user_type") == model.UserTypeClub {
		if comment.UserId == c.GetString("club_id") {
			comment.Right = true
		}
	} else {
		if comment.UserId == c.GetString("uid") {
			comment.Right = true
		}
	}
	return comment.Right
}

func CreateFeedsComment(c *gin.Context, feedId int64, req *forms.CreateFeedsCommentsRequest) (*model.FeedsComments, error) {
	_, err := QueryFeedInfo(feedId)
	if err != nil {
		global.Lg.Error("CreateFeedsComment error", zap.Error(err))
		return nil, err
	}

	userId := c.GetString("uid")
	userType := model.UserTypeUser
	if c.GetInt("user_type") == model.UserTypeClub {
		userType = model.UserTypeClub
		userId = c.GetString("club_id")
	}
	comment := &model.FeedsComments{
		UserId:   userId,
		UserType: userType,
		FeedId:   feedId,
		Content:  req.Content,
		Urls:     req.Urls,
	}

	if err := global.DB.Model(&model.FeedsComments{}).Create(&comment).Error; err != nil {
		global.Lg.Error("CreateFeedsComment error", zap.Error(err))
		return nil, err
	}

	return comment, nil
}

func QueryFeedsCommentInfo(c *gin.Context, feedId int64, id int64) (*model.FeedsComments, error) {
	comment := &model.FeedsComments{}
	if err := global.DB.Model(&model.FeedsComments{}).Where("id = ? and state = 0 and feed_id = ?", id, feedId).First(&comment).Error; err != nil {
		global.Lg.Error("QueryFeedsCommentInfo error", zap.Error(err))
		return nil, err
	}

	checkRight(c, comment)
	return comment, nil
}

func UpdateFeedsComment(c *gin.Context, feedId int64, id int64, req *forms.CreateFeedsCommentsRequest) (*model.FeedsComments, error) {
	comment, err := QueryFeedsCommentInfo(c, feedId, id)
	if err != nil {
		global.Lg.Error("UpdateFeedsComment error", zap.Error(err))
		return nil, err
	}

	if !comment.Right {
		return nil, enum.NewErr(enum.ParamErr, "无权限操作评论")
	}

	comment.Content = req.Content
	comment.Urls = req.Urls
	if err = global.DB.Model(&model.FeedsComments{}).Where("id = ? and state = 0", id).Updates(comment).Error; err != nil {
		global.Lg.Error("UpdateFeedsComment error", zap.Error(err))
		return nil, err
	}
	return comment, nil
}

func DeleteFeedsComment(c *gin.Context, feedId int64, id int64) error {
	comment, err := QueryFeedsCommentInfo(c, feedId, id)
	if err != nil {
		global.Lg.Error("DeleteFeedsComment error", zap.Error(err))
		return err
	}

	if !comment.Right {
		return enum.NewErr(enum.ParamErr, "无权限操作评论")
	}

	if err := global.DB.Model(&model.FeedsComments{}).Where("id = ? and state = 0", id).Updates(map[string]interface{}{"state": 1}).Error; err != nil {
		global.Lg.Error("DeleteFeedsComment error", zap.Error(err))
		return err
	}
	return nil
}
