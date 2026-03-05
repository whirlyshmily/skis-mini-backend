package forms

import "skis-admin-backend/model"

type OrdersCoursesComments struct {
	Content string             `json:"content"`
	Urls    model.JsonUrlArray `json:"urls"`
}
