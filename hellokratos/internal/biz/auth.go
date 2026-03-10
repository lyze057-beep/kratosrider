package biz

import (
	"context"
	"errors"
	"fmt"
	"hellokratos/internal/data"
	"hellokratos/internal/data/model"
	"hellokratos/internal/data/sms"
	"hellokratos/internal/pkg/jwt"
	"hellokratos/internal/pkg/wechat"
	"math/rand"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthUsecase 认证相关的业务逻辑接口
type AuthUsecase interface {
	// SendCode 发送验证码
	SendCode(ctx context.Context, phone string, codeType string) (string, string, int32, error)
	// Register 注册
	Register(ctx context.Context, phone string, password string, code string, nickname string) (*model.User, string, string, int64, error)
	// LoginByPhone 手机号+验证码登录
	LoginByPhone(ctx context.Context, phone string, code string) (*model.User, string, string, int64, error)
	// LoginByPassword 密码登录
	LoginByPassword(ctx context.Context, phone string, password string) (*model.User, string, string, int64, error)
	// Logout 登出
	Logout(ctx context.Context, userID string) error
	// RefreshToken 刷新token
	RefreshToken(ctx context.Context, refreshToken string) (*model.User, string, string, int64, error)
	// LoginByThirdParty 第三方登录
	LoginByThirdParty(ctx context.Context, platform string, code string) (*model.User, string, string, int64, error)
}

// authUsecase 认证相关的业务逻辑实现
type authUsecase struct {
	authRepo  data.AuthRepo
	rdb       *redis.Client
	sms       *sms.HywxSMS
	jwtMgr    *jwt.JWTManager
	wechatMgr *wechat.WeChatManager
	log       *log.Helper
}

// NewAuthUsecase 创建认证业务逻辑实例
func NewAuthUsecase(authRepo data.AuthRepo, rdb *redis.Client, sms *sms.HywxSMS, logger log.Logger) AuthUsecase {
	secretKey := "your-secret-key-change-this-in-production"
	tokenDuration := 24 * time.Hour
	jwtMgr := jwt.NewJWTManager(secretKey, tokenDuration)
	jwtMgr.SetRedis(rdb)

	// 微信配置，实际应用中应该从配置文件读取
	wechatAppID := "your-wechat-appid"
	wechatAppSecret := "your-wechat-appsecret"
	wechatMgr := wechat.NewWeChatManager(wechatAppID, wechatAppSecret)

	return &authUsecase{
		authRepo:  authRepo,
		rdb:       rdb,
		sms:       sms,
		jwtMgr:    jwtMgr,
		wechatMgr: wechatMgr,
		log:       log.NewHelper(logger),
	}
}

// SendCode 发送验证码
func (uc *authUsecase) SendCode(ctx context.Context, phone string, codeType string) (string, string, int32, error) {
	// 生成6位验证码
	code := generateCode(6)
	// 验证码有效期5分钟
	expireSeconds := int32(300)
	// 存储验证码到Redis
	key := fmt.Sprintf("sms:code:%s:%s", phone, codeType)
	uc.log.Info("storing code to redis", "key", key, "code", code, "expire", expireSeconds)
	err := uc.rdb.Set(ctx, key, code, time.Duration(expireSeconds)*time.Second).Err()
	if err != nil {
		uc.log.Error("failed to store code", "err", err)
		return "", "", 0, err
	}
	// 验证是否存储成功
	storedCode, err := uc.rdb.Get(ctx, key).Result()
	if err != nil {
		uc.log.Error("failed to verify code storage", "err", err)
	} else {
		uc.log.Info("code stored successfully", "stored_code", storedCode)
	}
	// 调用互亿无线短信发送API
	/*
		if uc.sms != nil {
			err := uc.sms.SendSMS(phone, code)
			if err != nil {
				uc.log.Error("failed to send sms", "err", err)
				// 短信发送失败不影响验证码的生成和存储，仍然返回成功
			}
			uc.log.Info("send code", "phone", phone, "code", code)
		} else {
			uc.log.Info("sms service not configured, skipping sms sending", "phone", phone, "code", code)
		}
	*/
	// 注释掉短信发送代码，仅用于测试
	uc.log.Info("send code (sms disabled for testing)", "phone", phone, "code", code)
	return "dummy_request_id", code, expireSeconds, nil
}

// Register 注册
func (uc *authUsecase) Register(ctx context.Context, phone string, password string, code string, nickname string) (*model.User, string, string, int64, error) {
	// 验证验证码
	key := fmt.Sprintf("sms:code:%s:register", phone)
	uc.log.Info("getting code from redis", "key", key)
	storedCode, err := uc.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			uc.log.Info("code not found in redis", "key", key)
			return nil, "", "", 0, errors.New("验证码已过期")
		}
		uc.log.Error("failed to get code from redis", "err", err)
		return nil, "", "", 0, err
	}
	uc.log.Info("code found in redis", "stored_code", storedCode, "input_code", code)
	if storedCode != code {
		uc.log.Info("code mismatch", "stored_code", storedCode, "input_code", code)
		return nil, "", "", 0, errors.New("验证码错误")
	}
	// 删除验证码
	uc.rdb.Del(ctx, key)
	uc.log.Info("code deleted from redis", "key", key)

	// 检查用户是否已存在
	existingUser, err := uc.authRepo.GetUserByPhone(ctx, phone)
	if err == nil {
		// 用户已存在，检查是否有密码
		if existingUser.Password != "" {
			uc.log.Info("user already exists with password", "phone", phone)
			return nil, "", "", 0, errors.New("用户已存在")
		}
		// 用户存在但没有密码，更新密码和昵称
		uc.log.Info("user exists without password, updating", "phone", phone)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			uc.log.Error("failed to hash password", "err", err)
			return nil, "", "", 0, err
		}
		existingUser.Password = string(hashedPassword)
		if nickname != "" {
			existingUser.Nickname = nickname
		}
		err = uc.authRepo.UpdateUser(ctx, existingUser)
		if err != nil {
			uc.log.Error("failed to update user", "err", err)
			return nil, "", "", 0, err
		}
		uc.log.Info("user updated with password", "user_id", existingUser.ID, "phone", existingUser.Phone)
		// 生成token和refresh token
		token, err := uc.jwtMgr.Generate(fmt.Sprintf("%d", existingUser.ID), existingUser.Phone, existingUser.Nickname)
		if err != nil {
			uc.log.Error("failed to generate token", "err", err)
			return nil, "", "", 0, err
		}
		refreshToken, err := uc.jwtMgr.Generate(fmt.Sprintf("%d", existingUser.ID), existingUser.Phone, existingUser.Nickname)
		if err != nil {
			uc.log.Error("failed to generate refresh token", "err", err)
			return nil, "", "", 0, err
		}
		expiresIn := int64(uc.jwtMgr.GetTokenDuration().Seconds())
		return existingUser, token, refreshToken, expiresIn, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		uc.log.Error("failed to get user", "err", err)
		return nil, "", "", 0, err
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		uc.log.Error("failed to hash password", "err", err)
		return nil, "", "", 0, err
	}

	// 创建用户
	user := &model.User{
		Phone:    phone,
		Password: string(hashedPassword),
		Nickname: nickname,
		Status:   0,
	}
	err = uc.authRepo.CreateUser(ctx, user)
	if err != nil {
		uc.log.Error("failed to create user", "err", err)
		return nil, "", "", 0, err
	}
	uc.log.Info("new user created", "user_id", user.ID, "phone", user.Phone)

	// 生成token和refresh token
	token, err := uc.jwtMgr.Generate(fmt.Sprintf("%d", user.ID), user.Phone, user.Nickname)
	if err != nil {
		uc.log.Error("failed to generate token", "err", err)
		return nil, "", "", 0, err
	}
	refreshToken, err := uc.jwtMgr.Generate(fmt.Sprintf("%d", user.ID), user.Phone, user.Nickname)
	if err != nil {
		uc.log.Error("failed to generate refresh token", "err", err)
		return nil, "", "", 0, err
	}
	expiresIn := int64(uc.jwtMgr.GetTokenDuration().Seconds())

	return user, token, refreshToken, expiresIn, nil
}

// LoginByPhone 手机号+验证码登录
func (uc *authUsecase) LoginByPhone(ctx context.Context, phone string, code string) (*model.User, string, string, int64, error) {
	// 验证验证码
	key := fmt.Sprintf("sms:code:%s:login", phone)
	uc.log.Info("getting code from redis", "key", key)
	storedCode, err := uc.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			uc.log.Info("code not found in redis", "key", key)
			return nil, "", "", 0, errors.New("验证码已过期")
		}
		uc.log.Error("failed to get code from redis", "err", err)
		return nil, "", "", 0, err
	}
	uc.log.Info("code found in redis", "stored_code", storedCode, "input_code", code)
	if storedCode != code {
		uc.log.Info("code mismatch", "stored_code", storedCode, "input_code", code)
		return nil, "", "", 0, errors.New("验证码错误")
	}
	// 删除验证码
	uc.rdb.Del(ctx, key)
	uc.log.Info("code deleted from redis", "key", key)

	// 查找用户
	user, err := uc.authRepo.GetUserByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 新用户，自动注册
			user = &model.User{
				Phone:    phone,
				Nickname: fmt.Sprintf("骑手%s", phone[len(phone)-4:]),
				Status:   0,
			}
			err = uc.authRepo.CreateUser(ctx, user)
			if err != nil {
				uc.log.Error("failed to create user", "err", err)
				return nil, "", "", 0, err
			}
			uc.log.Info("new user created", "user_id", user.ID, "phone", user.Phone)
		} else {
			uc.log.Error("failed to get user", "err", err)
			return nil, "", "", 0, err
		}
	} else {
		uc.log.Info("user found", "user_id", user.ID, "phone", user.Phone)
	}

	// 生成token和refresh token
	token, err := uc.jwtMgr.Generate(fmt.Sprintf("%d", user.ID), user.Phone, user.Nickname)
	if err != nil {
		uc.log.Error("failed to generate token", "err", err)
		return nil, "", "", 0, err
	}
	refreshToken, err := uc.jwtMgr.Generate(fmt.Sprintf("%d", user.ID), user.Phone, user.Nickname)
	if err != nil {
		uc.log.Error("failed to generate refresh token", "err", err)
		return nil, "", "", 0, err
	}
	expiresIn := int64(uc.jwtMgr.GetTokenDuration().Seconds())

	return user, token, refreshToken, expiresIn, nil
}

// LoginByPassword 密码登录
func (uc *authUsecase) LoginByPassword(ctx context.Context, phone string, password string) (*model.User, string, string, int64, error) {
	// 查找用户
	user, err := uc.authRepo.GetUserByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", "", 0, errors.New("用户不存在")
		}
		return nil, "", "", 0, err
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, "", "", 0, errors.New("密码错误")
	}

	// 生成token和refresh token
	token, err := uc.jwtMgr.Generate(fmt.Sprintf("%d", user.ID), user.Phone, user.Nickname)
	if err != nil {
		uc.log.Error("failed to generate token", "err", err)
		return nil, "", "", 0, err
	}
	refreshToken, err := uc.jwtMgr.Generate(fmt.Sprintf("%d", user.ID), user.Phone, user.Nickname)
	if err != nil {
		uc.log.Error("failed to generate refresh token", "err", err)
		return nil, "", "", 0, err
	}
	expiresIn := int64(uc.jwtMgr.GetTokenDuration().Seconds())

	return user, token, refreshToken, expiresIn, nil
}

// Logout 登出
func (uc *authUsecase) Logout(ctx context.Context, userID string) error {
	// 将用户的所有token加入黑名单
	// 这里简化处理，只将userID加入黑名单，实际应用中应该将具体的token加入黑名单
	key := fmt.Sprintf("token:blacklist:user:%s", userID)
	err := uc.rdb.Set(ctx, key, "1", uc.jwtMgr.GetTokenDuration()).Err()
	if err != nil {
		uc.log.Error("failed to add user to blacklist", "err", err)
		return err
	}
	uc.log.Info("user logged out", "user_id", userID)
	return nil
}

// RefreshToken 刷新token
func (uc *authUsecase) RefreshToken(ctx context.Context, refreshToken string) (*model.User, string, string, int64, error) {
	// 验证refresh token
	claims, err := uc.jwtMgr.Verify(ctx, refreshToken)
	if err != nil {
		uc.log.Error("failed to verify refresh token", "err", err)
		return nil, "", "", 0, errors.New("无效的refresh token")
	}

	// 获取用户信息
	userID := claims.UserID
	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		uc.log.Error("failed to parse user id", "err", err)
		return nil, "", "", 0, errors.New("无效的用户ID")
	}
	user, err := uc.authRepo.GetUserByID(ctx, userIDInt)
	if err != nil {
		uc.log.Error("failed to get user", "err", err)
		return nil, "", "", 0, errors.New("用户不存在")
	}

	// 生成新的token和refresh token
	newToken, err := uc.jwtMgr.Generate(fmt.Sprintf("%d", user.ID), user.Phone, user.Nickname)
	if err != nil {
		uc.log.Error("failed to generate new token", "err", err)
		return nil, "", "", 0, err
	}
	newRefreshToken, err := uc.jwtMgr.Generate(fmt.Sprintf("%d", user.ID), user.Phone, user.Nickname)
	if err != nil {
		uc.log.Error("failed to generate new refresh token", "err", err)
		return nil, "", "", 0, err
	}
	expiresIn := int64(uc.jwtMgr.GetTokenDuration().Seconds())

	uc.log.Info("token refreshed", "user_id", user.ID)

	return user, newToken, newRefreshToken, expiresIn, nil
}

// LoginByThirdParty 第三方登录
func (uc *authUsecase) LoginByThirdParty(ctx context.Context, platform string, code string) (*model.User, string, string, int64, error) {
	// 根据不同的平台执行不同的登录逻辑
	switch platform {
	case "wechat":
		// 微信登录逻辑
		return uc.loginByWechat(ctx, code)
	default:
		return nil, "", "", 0, errors.New("unsupported platform")
	}
}

// loginByWechat 微信登录
func (uc *authUsecase) loginByWechat(ctx context.Context, code string) (*model.User, string, string, int64, error) {
	// 使用code获取微信openid和unionid
	openID, unionID, err := uc.wechatMgr.GetUserInfoByCode(ctx, code)
	if err != nil {
		uc.log.Error("failed to get wechat user info", "err", err)
		return nil, "", "", 0, errors.New("获取微信用户信息失败")
	}

	// 使用unionID作为唯一标识，如果没有unionID则使用openID
	thirdPartyID := unionID
	if thirdPartyID == "" {
		thirdPartyID = openID
	}

	// 查找用户
	user, err := uc.authRepo.GetUserByThirdParty(ctx, "wechat", thirdPartyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 新用户，自动注册
			user = &model.User{
				ThirdPartyID:       thirdPartyID,
				ThirdPartyPlatform: "wechat",
				Nickname:           "微信用户",
				Status:             0,
			}
			err = uc.authRepo.CreateUser(ctx, user)
			if err != nil {
				uc.log.Error("failed to create user", "err", err)
				return nil, "", "", 0, err
			}
			uc.log.Info("new wechat user created", "user_id", user.ID, "openid", openID)
		} else {
			uc.log.Error("failed to get user", "err", err)
			return nil, "", "", 0, err
		}
	} else {
		uc.log.Info("wechat user found", "user_id", user.ID, "openid", openID)
	}

	// 生成token和refresh token
	token, err := uc.jwtMgr.Generate(fmt.Sprintf("%d", user.ID), user.Phone, user.Nickname)
	if err != nil {
		uc.log.Error("failed to generate token", "err", err)
		return nil, "", "", 0, err
	}
	refreshToken, err := uc.jwtMgr.Generate(fmt.Sprintf("%d", user.ID), user.Phone, user.Nickname)
	if err != nil {
		uc.log.Error("failed to generate refresh token", "err", err)
		return nil, "", "", 0, err
	}
	expiresIn := int64(uc.jwtMgr.GetTokenDuration().Seconds())

	return user, token, refreshToken, expiresIn, nil
}

// generateCode 生成指定长度的验证码
func generateCode(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := ""
	for i := 0; i < length; i++ {
		code += fmt.Sprintf("%d", r.Intn(10))
	}
	return code
}
