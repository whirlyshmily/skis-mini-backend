package forms

import "skis-admin-backend/model"

type QueryOrdersCoursesRecordsRequest struct {
	ListQueryRequest
	Uid           string `form:"uid"`
	OrderId       string `form:"order_id"`
	OrderCourseId string `form:"order_course_id"`
	GoodId        string `form:"good_id"`
	CourseId      string `form:"course_id"`
}

type CreateOrdersCoursesRecordItem struct {
	Content string             `json:"content" binding:"required"`
	Urls    model.JsonUrlArray `json:"urls"`
}

type CreateOrdersCoursesRecordRequest struct {
	Records []CreateOrdersCoursesRecordItem `json:"records" binding:"required"`
}
