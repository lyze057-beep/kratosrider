package service

import (
	"context"
	"fmt"
	"hellokratos/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// DebugService 调试服务
type DebugService struct {
	db  *gorm.DB
	log *log.Helper
}

// NewDebugService 创建调试服务
func NewDebugService(db *gorm.DB, logger log.Logger) *DebugService {
	return &DebugService{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

// ResetPasswordResponse 重置密码响应
type ResetPasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ResetPassword 重置密码（仅用于开发和测试）
func (s *DebugService) ResetPassword(ctx context.Context, req *ResetPasswordRequest) (*ResetPasswordResponse, error) {
	var user model.User
	if err := s.db.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
		return &ResetPasswordResponse{
			Success: false,
			Message: "用户不存在: " + req.Phone,
		}, nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return &ResetPasswordResponse{
			Success: false,
			Message: "密码加密失败",
		}, nil
	}

	user.Password = string(hashedPassword)
	if err := s.db.Save(&user).Error; err != nil {
		return &ResetPasswordResponse{
			Success: false,
			Message: "保存密码失败",
		}, nil
	}

	s.log.Info("password reset", "phone", req.Phone)
	return &ResetPasswordResponse{
		Success: true,
		Message: fmt.Sprintf("密码重置成功，手机号: %s，新密码: %s", req.Phone, req.Password),
	}, nil
}

// ListUsers 列出所有用户
func (s *DebugService) ListUsers(ctx context.Context) ([]model.User, error) {
	var users []model.User
	if err := s.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
