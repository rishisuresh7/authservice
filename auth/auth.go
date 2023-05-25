package auth

import (
	"context"
	"fmt"
	"time"

	"authservice/helper"
	"authservice/models"
	"authservice/repository"
)

type Authorizer interface {
	ValidateRefreshToken(ctx context.Context, token string) (map[string]interface{}, error)
	ValidateBearerToken(ctx context.Context, token string) (map[string]interface{}, error)
	GetJWT(ctx context.Context, claims map[string]interface{}, oldBearerToken, refreshToken string) (string, error)
	InvalidateTokens(ctx context.Context, userId, bearerToken string, clearAllTokens bool) error
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

	userMeta, err := a.GetActiveTokens(ctx, fmt.Sprintf("%s", refreshMeta.UserClaims["id"]))
	if err != nil {
		return nil, fmt.Errorf("validateRefreshToken: %s", err)
	}

	if !userMeta.ContainsRefreshToken(token) {
		return nil, fmt.Errorf("validateRefreshToken: not a valid token")
	}

	return refreshMeta.UserClaims, err
}

func (a *authorize) ValidateBearerToken(ctx context.Context, token string) (map[string]interface{}, error) {
	claims, err := a.helper.DecodeJWT(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %s", err)
	}

	userMeta, err := a.GetActiveTokens(ctx, claims["id"].(string))
	if err != nil {
		return nil, fmt.Errorf("validateBearerToken: %s", err)
	}

	if !userMeta.ContainsBearerToken(token) {
		return nil, fmt.Errorf("validateBearerToken: invalid token, token was invalidated")
	}

	return claims, nil
}

func (a *authorize) GetActiveTokens(ctx context.Context, userId string) (*models.UserMeta, error) {
	userMetaBytes, err := a.redis.GetBytes(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("getActiveTokens: unable to read token metadata")
	}

	var userMeta models.UserMeta
	a.helper.UnMarshal(userMetaBytes, &userMeta)

	return &userMeta, nil
}

func (a *authorize) UpdateActiveTokensWithNewBearerToken(ctx context.Context, userId, oldBearerToken, newBearerToken, refreshToken string) error {
	userMeta, err := a.GetActiveTokens(ctx, userId)
	if err != nil {
		return err
	}

	updated := userMeta.ReplaceBearerToken(oldBearerToken, newBearerToken, refreshToken)
	if !updated {
		return fmt.Errorf("no access/refresh token pair found")
	}

	err = a.redis.Set(ctx, userId, userMeta.GetBytes(), 0)
	if err != nil {
		return err
	}

	return nil
}

func (a *authorize) GetJWT(ctx context.Context, claims map[string]interface{}, oldBearerToken, refreshToken string) (string, error) {
	jwt, err := a.helper.GetJWT(claims)
	if err != nil {
		return "", fmt.Errorf("getJWT: unable to create JWT: %s", err)
	}

	err = a.UpdateActiveTokensWithNewBearerToken(ctx, claims["id"].(string), oldBearerToken, jwt, refreshToken)
	if err != nil {
		return "", fmt.Errorf("getJWT: %s", err)
	}

	return jwt, nil
}

func (a *authorize) InvalidateTokens(ctx context.Context, userId, bearerToken string, clearAllTokens bool) error {
	userMeta, err := a.GetActiveTokens(ctx, userId)
	if err != nil {
		return fmt.Errorf("invalidateTokens: %s", err)
	}

	if clearAllTokens {
		userMeta.ClearAllTokens()
	} else {
		userMeta.ClearToken(bearerToken)
	}

	err = a.redis.Set(ctx, userId, userMeta.GetBytes(), 0)
	if err != nil {
		return fmt.Errorf("invalidateTokens: unable to set data to redis: %s", err)
	}

	return nil
}