package data

import (
	"hellokratos/internal/data/model"
	"strings"

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
	// 自动迁移数据库表结构，忽略索引已存在的错误
	models := []interface{}{
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
		// 新增：骑手位置模块
		&model.RiderLocation{},
		&model.RiderLocationHistory{},
	}

	for _, m := range models {
		err := db.AutoMigrate(m)
		if err != nil {
			// 忽略索引已存在的错误
			if strings.Contains(err.Error(), "already exists") {
				continue
			}
			return err
		}
	}

	// 添加数据库索引（优化版）
	err := addIndexes(db)
	if err != nil {
		return err
	}

	return nil
}

// addIndexes 添加数据库索引（优化性能）
func addIndexes(db *gorm.DB) error {
	// 订单表索引
	indexes := []struct {
		table  string
		name   string
		column string
	}{
		// 订单表索引
		{"rider_order", "idx_order_status", "status"},
		{"rider_order", "idx_order_rider_id", "rider_id"},
		{"rider_order", "idx_order_status_rider_id", "status, rider_id"},
		{"rider_order", "idx_order_created_at", "created_at"},

		// 用户表索引
		{"rider_user", "idx_user_phone", "phone"},
		{"rider_user", "idx_user_status", "status"},
		{"rider_user", "idx_user_third_party", "third_party_platform, third_party_id"},

		// 位置表索引
		{"rider_location", "idx_location_rider_id", "rider_id"},
		{"rider_location", "idx_location_updated_at", "updated_at"},
		{"rider_location_history", "idx_history_rider_id", "rider_id"},
		{"rider_location_history", "idx_history_created_at", "created_at"},
	}

	for _, idx := range indexes {
		// 检查索引是否已存在
		var count int64
		db.Raw("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?", idx.table, idx.name).Scan(&count)

		if count == 0 {
			// 索引不存在，创建索引（使用字符串拼接，因为GORM不支持动态表名和列名）
			sql := "CREATE INDEX " + idx.name + " ON " + idx.table + " (" + idx.column + ")"
			err := db.Exec(sql).Error
			if err != nil {
				if strings.Contains(err.Error(), "already exists") {
					continue
				}
				return err
			}
		}
	}

	return nil
}
