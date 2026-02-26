package service

import (
	"context"

	pb "hellokratos/api/rider/v1"
)

type AuthService struct {
	pb.UnimplementedAuthServer
}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) LoginByPhone(ctx context.Context, req *pb.LoginByPhoneRequest) (*pb.LoginReply, error) {
    return &pb.LoginReply{}, nil
}
func (s *AuthService) LoginByPassword(ctx context.Context, req *pb.LoginByPasswordRequest) (*pb.LoginReply, error) {
    return &pb.LoginReply{}, nil
}
func (s *AuthService) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutReply, error) {
    return &pb.LogoutReply{}, nil
}
func (s *AuthService) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.LoginReply, error) {
    return &pb.LoginReply{}, nil
}
