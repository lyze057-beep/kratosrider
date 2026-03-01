package data

import (
	"context"
	"hellokratos/internal/data/model"

	"gorm.io/gorm"
)

// AuthRepo 认证相关的数据访问接口
type AuthRepo interface {
	// CreateUser 创建用户
	CreateUser(ctx context.Context, user *model.User) error
	// GetUserByPhone 根据手机号获取用户
	GetUserByPhone(ctx context.Context, phone string) (*model.User, error)
	// GetUserByID 根据ID获取用户
	GetUserByID(ctx context.Context, id int64) (*model.User, error)
	// GetUserByThirdParty 根据第三方平台和ID获取用户
	GetUserByThirdParty(ctx context.Context, platform string, thirdPartyID string) (*model.User, error)
	// UpdateUser 更新用户信息
	UpdateUser(ctx context.Context, user *model.User) error
}

// authRepo 认证相关的数据访问实现
type authRepo struct {
	db *gorm.DB
}

// NewAuthRepo 创建认证数据访问实例
func NewAuthRepo(data *Data) AuthRepo {
	return &authRepo{db: data.db}
}

// CreateUser 创建用户
func (r *authRepo) CreateUser(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetUserByPhone 根据手机号获取用户
func (r *authRepo) GetUserByPhone(ctx context.Context, phone string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID 根据ID获取用户
func (r *authRepo) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser 更新用户信息
func (r *authRepo) UpdateUser(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// GetUserByThirdParty 根据第三方平台和ID获取用户
func (r *authRepo) GetUserByThirdParty(ctx context.Context, platform string, thirdPartyID string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("third_party_platform = ? AND third_party_id = ?", platform, thirdPartyID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
