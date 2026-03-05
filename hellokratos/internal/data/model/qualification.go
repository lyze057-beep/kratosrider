package model

import (
	"time"
)

// RiderQualification 骑手资质验证表
type RiderQualification struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	RiderID         int64     `gorm:"uniqueIndex;not null" json:"rider_id"`
	OverallStatus   int32     `gorm:"default:0" json:"overall_status"` // 0-未认证 1-认证中 2-认证通过 3-认证失败
	IDCardStatus    int32     `gorm:"default:0" json:"id_card_status"` // 0-未提交 1-审核中 2-已通过 3-已拒绝
	LicenseStatus   int32     `gorm:"default:0" json:"license_status"` // 0-未提交 1-审核中 2-已通过 3-已拒绝
	HealthStatus    int32     `gorm:"default:0" json:"health_status"` // 0-未提交 1-审核中 2-已通过 3-已拒绝
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 表名
func (RiderQualification) TableName() string {
	return "rider_qualification"
}

// IDCardVerification 身份证验证表
type IDCardVerification struct {
	ID               int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	RiderID          int64     `gorm:"uniqueIndex;not null" json:"rider_id"`
	RealName         string    `gorm:"type:varchar(100);not null" json:"real_name"`
	IDCardNumber     string    `gorm:"type:varchar(18);not null" json:"id_card_number"`
	IDCardFront      string    `gorm:"type:varchar(500)" json:"id_card_front"`
	IDCardBack       string    `gorm:"type:varchar(500)" json:"id_card_back"`
	IDCardHandheld   string    `gorm:"type:varchar(500)" json:"id_card_handheld"`
	Status           int32     `gorm:"default:1" json:"status"` // 1-审核中 2-已通过 3-已拒绝
	VerifyResult     string    `gorm:"type:text" json:"verify_result"`
	RejectReason     string    `gorm:"type:varchar(500)" json:"reject_reason"`
	SubmitTime       time.Time `gorm:"autoCreateTime" json:"submit_time"`
	VerifyTime       *time.Time `json:"verify_time"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 表名
func (IDCardVerification) TableName() string {
	return "rider_idcard_verification"
}

// DriverLicenseVerification 驾驶证验证表
type DriverLicenseVerification struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	RiderID      int64     `gorm:"uniqueIndex;not null" json:"rider_id"`
	LicenseNumber string   `gorm:"type:varchar(50);not null" json:"license_number"`
	LicenseType  string    `gorm:"type:varchar(10);not null" json:"license_type"` // C1, C2, D, E等
	IssueDate    string    `gorm:"type:varchar(20)" json:"issue_date"`
	ExpiryDate   string    `gorm:"type:varchar(20)" json:"expiry_date"`
	LicenseFront string    `gorm:"type:varchar(500)" json:"license_front"`
	LicenseBack  string    `gorm:"type:varchar(500)" json:"license_back"`
	Status       int32     `gorm:"default:1" json:"status"` // 1-审核中 2-已通过 3-已拒绝
	VerifyResult string    `gorm:"type:text" json:"verify_result"`
	RejectReason string    `gorm:"type:varchar(500)" json:"reject_reason"`
	SubmitTime   time.Time `gorm:"autoCreateTime" json:"submit_time"`
	VerifyTime   *time.Time `json:"verify_time"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 表名
func (DriverLicenseVerification) TableName() string {
	return "rider_driver_license_verification"
}

// HealthCertificateVerification 健康证验证表
type HealthCertificateVerification struct {
	ID                int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	RiderID           int64     `gorm:"uniqueIndex;not null" json:"rider_id"`
	CertificateNumber string    `gorm:"type:varchar(50);not null" json:"certificate_number"`
	IssueDate         string    `gorm:"type:varchar(20)" json:"issue_date"`
	ExpiryDate        string    `gorm:"type:varchar(20)" json:"expiry_date"`
	CertificateImage  string    `gorm:"type:varchar(500)" json:"certificate_image"`
	HospitalName      string    `gorm:"type:varchar(200)" json:"hospital_name"`
	Status            int32     `gorm:"default:1" json:"status"` // 1-审核中 2-已通过 3-已拒绝
	VerifyResult      string    `gorm:"type:text" json:"verify_result"`
	RejectReason      string    `gorm:"type:varchar(500)" json:"reject_reason"`
	SubmitTime        time.Time `gorm:"autoCreateTime" json:"submit_time"`
	VerifyTime        *time.Time `json:"verify_time"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 表名
func (HealthCertificateVerification) TableName() string {
	return "rider_health_certificate_verification"
}
