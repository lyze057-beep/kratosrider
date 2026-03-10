package data

import (
	"hellokratos/internal/data/model"

	"gorm.io/gorm"
)

// AutoMigrate 自动迁移数据库表结构
//
// 参数:
//
//	db: 数据库连接
//
// 返回值:
//
//	error: 错误信息
func AutoMigrate(db *gorm.DB) error {
	// 自动迁移数据库表结构
	err := db.AutoMigrate(
		&model.User{},
		&model.Order{},
		&model.Message{},
		&model.Group{},
		&model.GroupMember{},
		&model.GroupMessage{},
		&model.Income{},
		&model.Withdrawal{},
		// 新增：AI智能体客服模块
		&model.AIAgentMessage{},
		&model.AIAgentFAQ{},
		&model.AIAgentSession{},
		// 新增：骑手资质验证模块
		&model.RiderQualification{},
		&model.IDCardVerification{},
		&model.DriverLicenseVerification{},
		&model.HealthCertificateVerification{},
		// 新增：骑手拉新模块
		&model.ReferralInviteCode{},
		&model.ReferralRelation{},
		&model.ReferralTask{},
		&model.ReferralTaskProgress{},
		&model.ReferralRewardRecord{},
		&model.ReferralRiskLog{},
		&model.ReferralStatistics{},
	)
	if err != nil {
		return err
	}

	return nil
}
