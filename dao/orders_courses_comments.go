package dao

import (
	"context"
	"errors"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func QueryOrdersCoursesComments(ctx context.Context, orderCourseId string) (*model.OrdersCoursesComments, error) {
	var comment *model.OrdersCoursesComments
	if err := global.DB.Model(&model.OrdersCoursesComments{}).
		Scopes(model.ScopeOrderCourseCommentFields).
		Preload("Reply", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(model.ScopeOrderCourseCommentFields)
		}).
		Preload("Reply.CoachInfo", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(model.ScopeCoachFields)
		}).
		Preload("Reply.CoachInfo.Users", func(db *gorm.DB) *gorm.DB {
			return model.ScopeUserSensitiveFields(db)
		}).
		Preload("UserInfo", func(db *gorm.DB) *gorm.DB {
			return model.ScopeUserSensitiveFields(db)
		}).
		Preload("Good").
		Preload("Good.CourseTags", "state = 0").
		Preload("Good.CourseTags.Tag", "state = 0").
		Where("order_course_id = ? and pid = 0 and state = 0", orderCourseId).Find(&comment).Error; err != nil {
		global.Lg.Error("查询订单课程评论失败", zap.Error(err))
		return nil, err
	}
	dealGoodTags(comment.Good)
	return comment, nil
}

func CreateOrdersCoursesComment(ctx context.Context, uid, orderCourseId string, req *forms.OrdersCoursesComments) (*model.OrdersCoursesComments, error) {
	orderCourse, err := OrderCourseInfo(ctx, uid, orderCourseId)
	if err != nil {
		global.Lg.Error("查询订单课程失败", zap.Error(err))
		return nil, err
	}

	//只能评论一次
	m, err := QueryOrdersCoursesComment(ctx, orderCourseId, uid)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		global.Lg.Error("查询订单课程评论失败", zap.Error(err))
		return nil, err
	}

	if m != nil {
		global.Lg.Error("订单课程已评论", zap.String("order_course_id", orderCourseId), zap.String("user_id", uid))
		return nil, errors.New("订单课程已评论")
	}

	comment := &model.OrdersCoursesComments{
		OrderId:       orderCourse.OrderID,
		OrderCourseId: orderCourseId,
		GoodId:        orderCourse.GoodID,
		CourseId:      orderCourse.CourseID,
		UserId:        uid,
		UserType:      model.UserTypeUser,
		Content:       req.Content,
		Urls:          req.Urls,
		OnShelf:       1,
	}

	if err = global.DB.Create(comment).Error; err != nil {
		global.Lg.Error("创建订单课程评论失败", zap.Error(err))
		return nil, err
	}

	return comment, nil
}

func CreateOrdersCoursesCommentReply(ctx context.Context, id int64, orderCourseId, coachId string, req *forms.OrdersCoursesComments) (*model.OrdersCoursesComments, error) {
	orderCourse, err := OrderCourseInfo(ctx, "", orderCourseId)
	if err != nil {
		global.Lg.Error("查询订单课程失败", zap.Error(err))
		return nil, err
	}

	comment, err := QueryOrdersCoursesCommentInfo(ctx, id)
	if err != nil {
		global.Lg.Error("查询订单课程评论失败", zap.Error(err))
		return nil, err
	}

	if comment.OrderCourseId != orderCourseId {
		global.Lg.Error("订单课程评论失败", zap.String("order_course_id", orderCourseId), zap.String("user_id", coachId))
		return nil, errors.New("订单课程评论失败")
	}

	//只能评论一次
	m, err := QueryOrdersCoursesCommentReply(ctx, orderCourseId, coachId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		global.Lg.Error("查询订单课程评论失败", zap.Error(err))
		return nil, err
	}

	if m != nil {
		global.Lg.Error("订单课程已评论", zap.String("order_course_id", orderCourseId), zap.String("user_id", coachId))
		return nil, errors.New("订单课程已回复")
	}

	comment = &model.OrdersCoursesComments{
		Pid:             id,
		OrderId:         orderCourse.OrderID,
		OrderCourseId:   orderCourseId,
		GoodId:          orderCourse.GoodID,
		CourseId:        orderCourse.CourseID,
		UserId:          coachId,
		UserType:        model.UserTypeCoach,
		RepliedUserId:   comment.UserId,
		RepliedUserType: comment.UserType,
		Content:         req.Content,
		Urls:            req.Urls,
		OnShelf:         1,
	}

	if err = global.DB.Create(comment).Error; err != nil {
		global.Lg.Error("创建订单课程评论失败", zap.Error(err))
		return nil, err
	}

	return comment, nil
}

func QueryOrdersCoursesComment(ctx context.Context, orderCourseId, userId string) (*model.OrdersCoursesComments, error) {
	var comment model.OrdersCoursesComments
	err := global.DB.Model(model.OrdersCoursesComments{}).Where("order_course_id = ? and user_id = ? and state = 0 and user_type = ?", orderCourseId, userId, model.UserTypeUser).First(&comment).Error
	if err != nil {
		global.Lg.Error("查询订单课程评论失败", zap.Error(err), zap.String("order_course_id", orderCourseId), zap.String("user_id", userId))
		return nil, err
	}
	return &comment, nil
}

func QueryOrdersCoursesCommentReply(ctx context.Context, orderCourseId, userId string) (*model.OrdersCoursesComments, error) {
	var comment model.OrdersCoursesComments
	err := global.DB.Model(model.OrdersCoursesComments{}).Where("order_course_id = ? and user_id = ? and state = 0 and user_type = ?", orderCourseId, userId, model.UserTypeCoach).First(&comment).Error
	if err != nil {
		global.Lg.Error("查询订单课程评论失败", zap.Error(err), zap.String("order_course_id", orderCourseId), zap.String("user_id", userId))
		return nil, err
	}
	return &comment, nil
}

func QueryOrdersCoursesCommentInfo(ctx context.Context, id int64) (*model.OrdersCoursesComments, error) {
	var comment model.OrdersCoursesComments
	err := global.DB.Model(model.OrdersCoursesComments{}).Where("id = ? and state = 0", id).First(&comment).Error
	if err != nil {
		global.Lg.Error("查询订单课程评论失败", zap.Error(err), zap.Int64("id", id))
		return nil, err
	}
	return &comment, nil
}

func QueryUserComments(ctx context.Context, userId string, userType int, req *forms.ListQueryRequest) (int64, []*model.OrdersCoursesComments, error) {
	db := global.DB.Model(&model.OrdersCoursesComments{}).
		Preload("Reply", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(model.ScopeOrderCourseCommentFields)
		}).
		Preload("Good").
		Preload("Good.CourseTags", "state = 0").
		Preload("Good.CourseTags.Tag", "state = 0").
		Preload("Comment", "state = 0").
		Preload("Comment.UserInfo", func(db *gorm.DB) *gorm.DB {
			return model.ScopeUserSensitiveFields(db)
		}).
		Where("user_id = ? and user_type = ? and state = 0", userId, userType)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		global.Lg.Error("查询用户课程评论失败", zap.Error(err))
		return 0, nil, err
	}

	var comments []*model.OrdersCoursesComments
	if err := db.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&comments).Error; err != nil {
		global.Lg.Error("查询教练课程评论失败", zap.Error(err))
		return 0, nil, err
	}
	for _, comment := range comments {
		dealGoodTags(comment.Good)
	}

	return total, comments, nil
}
