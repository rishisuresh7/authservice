package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"authservice/helper"
	"authservice/models"
	"authservice/repository"
)

type Authorizer interface {
	ValidateRefreshToken(ctx context.Context, token string) (map[string]interface{}, error)
	DecodeJWT(token string) (int64, error)
	GetJWT(ctx context.Context, claims map[string]interface{}) (string, error)
	InvalidateTokens(ctx context.Context, userId int64, userType string) error
}

type authorize struct {
	helper helper.Helper
	redis  repository.RedisQueryer
}

func NewAuthorizer(h helper.Helper, r repository.RedisQueryer) Authorizer {
	return &authorize{
		helper: h,
		redis: r,
	}
}

func (a *authorize) ValidateRefreshToken(ctx context.Context, token string) (map[string]interface{}, error) {
	refreshMeta, err := a.helper.DecodeToken(token)
	if err != nil {
		return nil, fmt.Errorf("validateRefreshToken: unable to validate token: %s", err)
	}

	if refreshMeta.Expiry <= time.Now().Unix() {
		return nil, fmt.Errorf("validateRefreshToken: invalid token: token expired")
	}

	userKey := fmt.Sprintf("%s-%.0f", refreshMeta.UserClaims["userType"], refreshMeta.UserClaims["id"].(float64))
	bytes, err := a.redis.GetBytes(ctx, userKey)
	if err != nil {
		return nil, fmt.Errorf("validateRefreshToken: unable to get data from redis: %s", err)
	}

	var userMeta models.UserMeta
	a.helper.UnMarshal(bytes, &userMeta)
	if userMeta.RefreshToken != token {
		return nil, fmt.Errorf("validateRefreshToken: expired refresh token")
	}

	return refreshMeta.UserClaims, err
}

func (a *authorize) GetJWT(ctx context.Context, claims map[string]interface{}) (string, error) {
	jwt, err := a.helper.GetJWT(claims)
	if err != nil {
		return "", fmt.Errorf("getJWT: unable to create JWT: %s", err)
	}

	userKey := fmt.Sprintf("%s-%.0f", claims["userType"], claims["id"].(float64))
	bytes, err := a.redis.GetBytes(ctx, userKey)
	if err != nil {
		return "", fmt.Errorf("getJWT: unable to get data from redis: %s", err)
	}

	var userMeta models.UserMeta
	a.helper.UnMarshal(bytes, &userMeta)
	userMeta.BearerToken = jwt
	err = a.redis.Set(ctx, userKey, userMeta.GetBytes(), 0)
	if err != nil {
		return "", fmt.Errorf("getJWT: unable to set data to redis: %s", err)
	}

	return jwt, err
}

func (a *authorize) InvalidateTokens(ctx context.Context, userId int64, userType string) error {
	userKey := fmt.Sprintf("%s-%d", userType, userId)
	bytes, err := a.redis.GetBytes(ctx, userKey)
	if err != nil {
		return fmt.Errorf("invalidateTokens: unable to get data from redis: %s", err)
	}

	var userMeta models.UserMeta
	a.helper.UnMarshal(bytes, &userMeta)
	userMeta.BearerToken = ""
	userMeta.RefreshToken = ""
	err = a.redis.Set(ctx, userKey, userMeta.GetBytes(), 0)
	if err != nil {
		return fmt.Errorf("invalidateTokens: unable to set data to redis: %s", err)
	}

	return nil
}

func (a *authorize) DecodeJWT(token string) (int64, error) {
	tokenString := strings.Split(token, "Bearer ")
	if len(tokenString) != 2 {
		return -1, fmt.Errorf("decodeJWT: invalid token format")
	}

	claims, err := a.helper.DecodeJWT(tokenString[1])
	if err != nil {
		if !strings.Contains(err.Error(), "Token is expired") {
			return -1, fmt.Errorf("decodeJWT: unable to decode JWT: %s", err)
		}
	}

	userId, ok := claims["id"].(float64)
	if !ok {
		return -1, fmt.Errorf("decodeJWT: invalid userId")
	}

	return int64(userId), nil
}