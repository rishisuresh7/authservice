package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"authservice/builder"
	"authservice/constant"
	"authservice/helper"
	"authservice/models"
	"authservice/repository"
)

type User interface {
	Register(ctx context.Context, user *models.User) (*models.User, error)
	GetUser(ctx context.Context, id, phone, email string) (bool, *models.User, error)
	Login(ctx context.Context, user *models.LoginUser) (*models.AuthUser, error)
	LoginWithOTP(ctx context.Context, user *models.LoginUser) (string, error)
	VerifyOTP(ctx context.Context, user *models.LoginUser) (bool, error)
	IsDeactivated(ctx context.Context, user *models.User) (bool, error)
	OAuthLogin(ctx context.Context, user models.User) (*models.AuthUser, error)
	ChangePassword(ctx context.Context, id string, cpr *models.ChangePasswordRequest) error
	ResetPassword(ctx context.Context, cpr *models.ChangePasswordRequest) error
	GetResetSecret(ctx context.Context, phone string) (string, error)
	UpdateUser(ctx context.Context, user *models.User) (*models.User, error)
}

type user struct {
	builder  builder.UserBuilder
	postgres repository.PostgresQueryer
	redis    repository.RedisQueryer
	helper   helper.Helper
}

func NewUser(b builder.UserBuilder, p repository.PostgresQueryer, r repository.RedisQueryer, h helper.Helper) User {
	return &user{
		builder:  b,
		postgres: p,
		redis:    r,
		helper:   h,
	}
}

func (u *user) IsDeactivated(ctx context.Context, user *models.User) (bool, error) {
	return false, nil
}

func (u *user) Login(ctx context.Context, user *models.LoginUser) (*models.AuthUser, error) {
	query := u.builder.Login(user.LoginType)
	res, err := u.postgres.QueryScan(ctx, query, user.Email, user.Phone, u.helper.Hash(user.Password))
	if err != nil {
		return nil, fmt.Errorf("login: unable to query data: %s", err)
	}

	defer res.Close()
	if !res.Next() {
		return nil, fmt.Errorf("login: user not found")
	}

	var us models.User
	err = res.Scan(&us)
	if err != nil {
		return nil, fmt.Errorf("login: unable to decode user: %s", err)
	}

	claims := map[string]interface{}{
		"id":        us.GetId(),
		"name": fmt.Sprintf("%s %s", us.GetFirstName(), us.GetLastName()),
	}
	token, err := u.helper.GetJWT(claims)
	if err != nil {
		return nil, fmt.Errorf("login: unable to get JWT: %s", err)
	}

	refreshToken, err := u.helper.EncodeClaims(claims)
	if err != nil {
		return nil, fmt.Errorf("login: unable to get refresh token: %s", err)
	}

	err = u.UpdateActiveTokens(ctx, us.GetId(), token, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("login: unable to store user meta: %s", err)
	}

	return &models.AuthUser{User: &us, BearerToken: token, RefreshToken: refreshToken}, nil
}

func (u *user) LoginWithOTP(ctx context.Context, user *models.LoginUser) (string, error) {
	exists, _, err := u.GetUser(ctx, "", user.Email, user.Phone)
	if err != nil {
		return "", fmt.Errorf("LoginWithOTP: %s", err)
	}

	if !exists {
		return "", fmt.Errorf("LoginWithOTP: user not registered")
	}

	nonce, err := u.helper.SendOTP(ctx, user.Phone)
	if err != nil {
		return "", fmt.Errorf("LoginWithOTP: unable to send OTP: %s", err)
	}

	return nonce, nil
}

func (u *user) VerifyOTP(ctx context.Context, user *models.LoginUser) (bool, error) {
	savedOtp, err := u.redis.GetString(ctx, user.Nonce)
	if err != nil {
		return false, fmt.Errorf("VerifyOTP: unable to get OTP: %s", err)
	}

	if savedOtp != user.OTP {
		return false, nil
	}

	return true, nil
}

func (u *user) GetUser(ctx context.Context, id, email, phone string) (bool, *models.User, error) {
	query := u.builder.GetUser()
	res, err := u.postgres.QueryScan(ctx, query, id, email, phone)
	if err != nil {
		return false, nil, fmt.Errorf("GetUser: unable to fetch user: %s", err)
	}

	defer res.Close()
	var user models.User
	if !res.Next() {
		return false, nil, nil
	}

	err = res.Scan(&user)
	if err != nil {
		return true, nil, fmt.Errorf("GetUser: unable to decode user: %s", err)
	}

	return true, &user, nil
}

func (u *user) Register(ctx context.Context, user *models.User) (*models.User, error) {
	id := u.helper.NewId()
	user.Id = &id
	query := u.builder.Register(user.GetMap())
	_, err := u.postgres.Exec(ctx, query, user.GetDOB(), u.helper.Hash(user.GetPassword()))
	if err != nil {
		if !strings.Contains(err.Error(), "duplicate key") {
			return nil, fmt.Errorf("register: unable to save data: %s", err)
		}
	}

	user.Password = nil
	user.ConfirmPassword = ""

	return user, nil
}

func (u *user) OAuthLogin(ctx context.Context, user models.User) (*models.AuthUser, error) {
	query := u.builder.Login("")
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

	err = u.UpdateActiveTokens(ctx, user.GetId(), token, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("oAuthLogin: unable to store user meta: %s", err)
	}

	return &models.AuthUser{User: &user, BearerToken: token, RefreshToken: refreshToken}, nil
}

func (u *user) ChangePassword(ctx context.Context, id string, cpr *models.ChangePasswordRequest) error {
	query := u.builder.ChangePassword()
	res, err := u.postgres.Exec(ctx, query, u.helper.Hash(cpr.NewPassword), u.helper.Hash(cpr.CurrentPassword), id)
	if err != nil {
		return fmt.Errorf("changePassword: unable to execute query: %s", err)
	}

	if res == 0 {
		return fmt.Errorf("no such user to update password")
	}

	return nil
}

func (u *user) ResetPassword(ctx context.Context, cpr *models.ChangePasswordRequest) error {
	phone, err := u.redis.GetString(ctx, cpr.Nonce)
	if err != nil {
		return fmt.Errorf("resetPassword: unable to get phone using nonce: %s", err)
	}

	fmt.Println(phone)
	query := u.builder.ResetPassword()
	res, err := u.postgres.Exec(ctx, query, u.helper.Hash(cpr.NewPassword), phone)
	if err != nil {
		return fmt.Errorf("resetPassword: unable to execute query: %s", err)
	}

	if res == 0 {
		return fmt.Errorf("no such user to reset password")
	}

	_, _ = u.redis.GetDelString(ctx, cpr.Nonce)
	return nil
}

func (u *user) GetResetSecret(ctx context.Context, phone string) (string, error) {
	nonce := u.helper.NewId()
	err := 	u.redis.Set(ctx, nonce, phone, 2*time.Minute)
	if err != nil {
		return "", fmt.Errorf("GetResetSecret: unable to store secret: %s", err)
	}

	return nonce, nil
}

func (u *user) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	query := u.builder.UpdateUser(user.GetMap())
	res, err := u.postgres.Exec(ctx, query, *user.Id)
	if err != nil {
		return nil, fmt.Errorf("UpdateUser: unable to execute query: %s", err)
	}

	if res == 0 {
		return nil, fmt.Errorf("UpdateUser: no such user to update")
	}

	_, usr, _ := u.GetUser(ctx, user.GetId(), "", "")
	return usr, nil
}

func (u *user) UpdateActiveTokens(ctx context.Context, userId, bearerToken, refreshToken string) error {
	var userMeta models.UserMeta
	userMetaBytes, err := u.redis.GetBytes(ctx, userId)
	if u.redis.IsRedisNil(err) {
		userMeta = models.UserMeta{
			UserId:        userId,
			LastLoginTime: time.Now().UnixMilli(),
			ActiveTokens:  []models.ActiveToken{
				{
					BearerToken: bearerToken,
					RefreshToken: refreshToken,
				},
			},
		}
	} else if err != nil {
		return fmt.Errorf("updateActiveTokens: unable to get redis key: %s", err)
	} else {
		u.helper.UnMarshal(userMetaBytes, &userMeta)
		userMeta.AddToken(bearerToken, refreshToken)
	}

	err = u.redis.Set(ctx, userId, userMeta.GetBytes(), 0)
	if err != nil {
		return fmt.Errorf("updateActiveTokens: unable to save to redis: %s", err)
	}

	return nil
}