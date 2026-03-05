package dao

import (
	"go.uber.org/zap"
	"skis-admin-backend/forms"
	"skis-admin-backend/global"
	"skis-admin-backend/model"
)

func CreateCertificateConfigs(req forms.CreateCertificateConfigRequest) (*model.CertificateConfigs, error) {
	config := &model.CertificateConfigs{
		Name:  req.Name,
		Level: req.Level,
	}
	if err := global.DB.Create(config).Error; err != nil {
		global.Lg.Error("创建证书标签失败", zap.Error(err))
		return nil, err
	}
	return config, nil
}

func QueryCertificateConfigsList(req forms.QueryCertificateConfigsListRequest) ([]*model.CertificateConfigs, error) {
	var configs []*model.CertificateConfigs
	db := global.DB.Model(&model.CertificateConfigs{}).Where("state", 0)
	if req.Keyword != "" {
		db = db.Where("name LIKE ?", "%"+req.Keyword+"%")
	}

	if err := db.Order("id desc").Find(&configs).Error; err != nil {
		global.Lg.Error("查询证书标签列表失败", zap.Error(err))
		return nil, err
	}
	return configs, nil
}

func QueryCertificateConfigInfo(id int64) (*model.CertificateConfigs, error) {
	var config model.CertificateConfigs
	if err := global.DB.Model(&model.CertificateConfigs{}).Where("id = ? and state = 0", id).First(&config).Error; err != nil {
		global.Lg.Error("查询证书标签详情失败", zap.Error(err))
		return nil, err
	}
	return &config, nil
}

func UpdateCertificateConfig(id int64, req forms.CreateCertificateConfigRequest) (*model.CertificateConfigs, error) {
	config, err := QueryCertificateConfigInfo(id)
	if err != nil {
		global.Lg.Error("查询证书标签详情失败", zap.Error(err))
		return nil, err
	}

	config.Name = req.Name
	config.Level = req.Level
	if err = global.DB.Model(&model.CertificateConfigs{}).Where("id = ? and state = 0", id).Updates(config).Error; err != nil {
		global.Lg.Error("更新证书标签失败", zap.Error(err))
		return nil, err
	}

	return config, nil
}

func DeleteCertificateConfig(id int64) error {
	_, err := QueryCertificateConfigInfo(id)
	if err != nil {
		global.Lg.Error("查询证书标签详情失败", zap.Error(err))
		return err
	}

	//todo 有引用不能删除
	if err = global.DB.Model(&model.CertificateConfigs{}).Where("id = ? and state = 0", id).Update("state", 1).Error; err != nil {
		global.Lg.Error("删除证书标签失败", zap.Error(err))
		return err
	}
	return nil
}
