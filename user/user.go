package user

import (
	"context"
	"fmt"
	"strings"

	"authservice/builder"
	"authservice/constant"
	"authservice/helper"
	"authservice/models"
	"authservice/repository"
)

type User interface {
	Register(ctx context.Context, phone string) (string, error)
	Verify(ctx context.Context, user *models.RegisterUser) (bool, error)
	Login(ctx context.Context, user *models.RegisterUser) (*models.AuthUser, error)
	IsDeactivated(ctx context.Context, user *models.RegisterUser) (bool, error)
	OAuthLogin(ctx context.Context, user models.User) (*models.AuthUser, error)
}

type user struct {
	builder  builder.UserBuilder
	postgres repository.PostgresQueryer
	redis    repository.RedisQueryer
	helper   helper.Helper
}

func NewCustomer(b builder.UserBuilder, p repository.PostgresQueryer, r repository.RedisQueryer, h helper.Helper) User {
	return &user{
		builder:  b,
		postgres: p,
		redis:    r,
		helper:   h,
	}
}

func (u *user) IsDeactivated(ctx context.Context, user *models.RegisterUser) (bool, error) {
	return false, nil
}

func (u *user) Login(ctx context.Context, user *models.RegisterUser) (*models.AuthUser, error) {
	query := u.builder.Login("phone")
	res, err := u.postgres.QueryScan(ctx, query, user.Phone)
	if err != nil {
		return nil, fmt.Errorf("login: unable to query data: %s", err)
	}

	defer res.Close()
	if !res.Next() {
		return nil, fmt.Errorf("login: customer not found")
	}

	var us models.User
	err = res.Scan(&us)
	if err != nil {
		return nil, fmt.Errorf("login: unable to decode user: %s", err)
	}

	claims := map[string]interface{}{
		"id":        *us.Id,
		"firstName": us.FirstName,
		"userType":  constant.User,
	}
	token, err := u.helper.GetJWT(claims)
	if err != nil {
		return nil, fmt.Errorf("login: unable to get JWT: %s", err)
	}

	refreshToken, err := u.helper.EncodeClaims(claims)
	if err != nil {
		return nil, fmt.Errorf("login: unable to get refresh token: %s", err)
	}

	userMeta := &models.UserMeta{
		UserId:       *us.Id,
		UserType:     constant.User,
		BearerToken:  token,
		RefreshToken: refreshToken,
	}
	err = u.redis.Set(ctx, fmt.Sprintf("%s-%s", constant.User, *us.Id), userMeta.GetBytes(), 0)
	if err != nil {
		return nil, fmt.Errorf("login: unable to store user meta: %s", err)
	}

	return &models.AuthUser{User: &us, BearerToken: token, RefreshToken: refreshToken}, nil
}

func (u *user) Register(ctx context.Context, phone string) (string, error) {
	query := u.builder.Register()
	_, err := u.postgres.Exec(ctx, query, phone, u.helper.NewId())
	if err != nil {
		if !strings.Contains(err.Error(), "duplicate key") {
			return "", fmt.Errorf("register: unable to save data: %s", err)
		}
	}

	nonce, err := u.helper.SendOTP(ctx, phone)
	if err != nil {
		return "", fmt.Errorf("register: unable to send OTP: %s", err)
	}

	return nonce, nil
}

func (u *user) Verify(ctx context.Context, user *models.RegisterUser) (bool, error) {
	savedOtp, err := u.redis.GetDelString(ctx, user.Nonce)
	if err != nil {
		return false, fmt.Errorf("verify: unable to verify: %s", err)
	}

	if savedOtp != user.OTP {
		return true, nil
	}

	return false, nil
}

func (u *user) OAuthLogin(ctx context.Context, user models.User) (*models.AuthUser, error) {
	query := u.builder.Login("email")
	res, err := u.postgres.QueryScan(ctx, query, user.Email)
	if err != nil {
		return nil, fmt.Errorf("oAuthLogin: unable to query data: %s", err)
	}

	if !res.Next() {
		query := u.builder.OAuthRegister()
		res, err := u.postgres.QueryScan(ctx, query, user.Email, user.FirstName, user.LastName)
		if err != nil {
			return nil, fmt.Errorf("oAuthLogin: unable to register user: %s", err)
		}

		if res.Next() {
			err = res.Scan(&user)
			res.Close()
			if err != nil {
				return nil, fmt.Errorf("oAuthLogin: unable to decode user: %s", err)
			}
		}
	} else {
		err = res.Scan(&user)
		res.Close()
		if err != nil {
			return nil, fmt.Errorf("oAuthLogin: unable to decode user: %s", err)
		}
	}

	claims := map[string]interface{}{
		"id":        user.Id,
		"firstName": user.FirstName,
		"userType":  constant.User,
	}
	token, err := u.helper.GetJWT(claims)
	if err != nil {
		return nil, fmt.Errorf("oAuthLogin: unable to get JWT: %s", err)
	}

	refreshToken, err := u.helper.EncodeClaims(claims)
	if err != nil {
		return nil, fmt.Errorf("oAuthLogin: unable to get refresh token: %s", err)
	}

	userMeta := &models.UserMeta{
		UserId:       *user.Id,
		UserType:     constant.User,
		BearerToken:  token,
		RefreshToken: refreshToken,
	}
	err = u.redis.Set(ctx, fmt.Sprintf("%s-%s", constant.User, *user.Id), userMeta.GetBytes(), 0)
	if err != nil {
		return nil, fmt.Errorf("oAuthLogin: unable to store user meta: %s", err)
	}

	return &models.AuthUser{User: &user, BearerToken: token, RefreshToken: refreshToken}, nil
}
