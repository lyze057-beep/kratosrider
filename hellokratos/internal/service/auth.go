package service

import (
	"context"
	"fmt"

	v1 "hellokratos/api/rider/v1"
	"hellokratos/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// AuthService 认证服务
type AuthService struct {
	v1.UnimplementedAuthServer
	authUsecase biz.AuthUsecase
	log         *log.Helper
}

// NewAuthService 创建认证服务实例
func NewAuthService(authUsecase biz.AuthUsecase, logger log.Logger) *AuthService {
	return &AuthService{
		authUsecase: authUsecase,
		log:         log.NewHelper(logger),
	}
}

// Register 注册
func (s *AuthService) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.LoginReply, error) {
	user, token, refreshToken, expiresIn, err := s.authUsecase.Register(ctx, req.Phone, req.Password, req.Code, req.Nickname)
	if err != nil {
		s.log.Error("register failed", "err", err)
		return nil, err
	}
	return &v1.LoginReply{
		UserId:       fmt.Sprintf("%d", user.ID),
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		UserInfo: &v1.UserInfo{
			UserId:   fmt.Sprintf("%d", user.ID),
			Phone:    user.Phone,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Status:   user.Status,
		},
	}, nil
}

// LoginByPhone 手机号+验证码登录
func (s *AuthService) LoginByPhone(ctx context.Context, req *v1.LoginByPhoneRequest) (*v1.LoginReply, error) {
	user, token, refreshToken, expiresIn, err := s.authUsecase.LoginByPhone(ctx, req.Phone, req.Code)
	if err != nil {
		s.log.Error("login by phone failed", "err", err)
		return nil, err
	}
	return &v1.LoginReply{
		UserId:       fmt.Sprintf("%d", user.ID),
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		UserInfo: &v1.UserInfo{
			UserId:   fmt.Sprintf("%d", user.ID),
			Phone:    user.Phone,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Status:   user.Status,
		},
	}, nil
}

// LoginByPassword 密码登录
func (s *AuthService) LoginByPassword(ctx context.Context, req *v1.LoginByPasswordRequest) (*v1.LoginReply, error) {
	user, token, refreshToken, expiresIn, err := s.authUsecase.LoginByPassword(ctx, req.Phone, req.Password)
	if err != nil {
		s.log.Error("login by password failed", "err", err)
		return nil, err
	}
	return &v1.LoginReply{
		UserId:       fmt.Sprintf("%d", user.ID),
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		UserInfo: &v1.UserInfo{
			UserId:   fmt.Sprintf("%d", user.ID),
			Phone:    user.Phone,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Status:   user.Status,
		},
	}, nil
}

// Logout 登出
func (s *AuthService) Logout(ctx context.Context, req *v1.LogoutRequest) (*v1.LogoutReply, error) {
	err := s.authUsecase.Logout(ctx, req.UserId)
	if err != nil {
		s.log.Error("logout failed", "err", err)
		return nil, err
	}
	return &v1.LogoutReply{
		Success: true,
	}, nil
}

// RefreshToken 刷新token
func (s *AuthService) RefreshToken(ctx context.Context, req *v1.RefreshTokenRequest) (*v1.LoginReply, error) {
	user, token, refreshToken, expiresIn, err := s.authUsecase.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		s.log.Error("refresh token failed", "err", err)
		return nil, err
	}
	return &v1.LoginReply{
		UserId:       fmt.Sprintf("%d", user.ID),
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		UserInfo: &v1.UserInfo{
			UserId:   fmt.Sprintf("%d", user.ID),
			Phone:    user.Phone,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Status:   user.Status,
		},
	}, nil
}

// SendCode 发送验证码
func (s *AuthService) SendCode(ctx context.Context, req *v1.SendCodeRequest) (*v1.SendCodeReply, error) {
	_, expireSeconds, err := s.authUsecase.SendCode(ctx, req.Phone, req.Type)
	if err != nil {
		s.log.Error("send code failed", "err", err)
		return nil, err
	}
	return &v1.SendCodeReply{
		RequestId:     "dummy_request_id",
		ExpireSeconds: expireSeconds,
	}, nil
}

// LoginByThirdParty 第三方登录
func (s *AuthService) LoginByThirdParty(ctx context.Context, req *v1.LoginByThirdPartyRequest) (*v1.LoginReply, error) {
	user, token, refreshToken, expiresIn, err := s.authUsecase.LoginByThirdParty(ctx, req.Platform, req.Code)
	if err != nil {
		s.log.Error("login by third party failed", "err", err)
		return nil, err
	}
	return &v1.LoginReply{
		UserId:       fmt.Sprintf("%d", user.ID),
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		UserInfo: &v1.UserInfo{
			UserId:   fmt.Sprintf("%d", user.ID),
			Phone:    user.Phone,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Status:   user.Status,
		},
	}, nil
}
