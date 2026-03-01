package wechat

import (
	"context"
	"fmt"
)

type WeChatManager struct {
	appID     string
	appSecret string
}

func NewWeChatManager(appID string, appSecret string) *WeChatManager {
	return &WeChatManager{
		appID:     appID,
		appSecret: appSecret,
	}
}

func (m *WeChatManager) GetAccessToken(ctx context.Context) (string, error) {
	// TODO: 实现获取access token的逻辑
	// 实际应用中需要调用微信API
	return "", fmt.Errorf("not implemented")
}

func (m *WeChatManager) GetUserInfoByCode(ctx context.Context, code string) (string, string, error) {
	// TODO: 实现通过code获取用户信息的逻辑
	// 这里返回模拟数据，实际应用中需要调用微信API
	openID := "mock_openid_" + code
	unionID := "mock_unionid_" + code
	return openID, unionID, nil
}
