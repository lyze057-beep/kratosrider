package jwt

import (
	"context"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token expired")
	ErrTokenBlacklisted = errors.New("token is blacklisted")
)

type Claims struct {
	UserID   string `json:"user_id"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secretKey     []byte
	tokenDuration time.Duration
	rdb           *redis.Client
	log           *log.Helper
}

func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		tokenDuration: tokenDuration,
		log:           log.NewHelper(log.DefaultLogger),
	}
}

func (manager *JWTManager) SetRedis(rdb *redis.Client) {
	manager.rdb = rdb
}

func (manager *JWTManager) Generate(userID string, phone string, nickname string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Phone:    phone,
		Nickname: nickname,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(manager.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(manager.secretKey)
}

func (manager *JWTManager) Verify(ctx context.Context, accessToken string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, ErrInvalidToken
			}
			return manager.secretKey, nil
		},
	)

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// 检查token是否在黑名单中
	if manager.rdb != nil {
		key := "token:blacklist:user:" + claims.UserID
		_, err := manager.rdb.Get(ctx, key).Result()
		if err == nil {
			// token在黑名单中
			return nil, ErrTokenBlacklisted
		}
	}

	return claims, nil
}

func (manager *JWTManager) GetTokenDuration() time.Duration {
	return manager.tokenDuration
}
