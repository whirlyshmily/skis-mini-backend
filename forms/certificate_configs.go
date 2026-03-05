package forms

import "skis-admin-backend/model"

type CreateCertificateConfigRequest struct {
	Name  string          `json:"name" binding:"required"`
	Level model.JSONArray `json:"level" binding:"required"`
}

type QueryCertificateConfigsListRequest struct {
	Keyword string `form:"keyword"`
}

type QueryCertificateConfigsListResponse struct {
	List []*model.CertificateConfigs `json:"list"`
}
