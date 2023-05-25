package models

import (
	"fmt"
	"strings"
	"unicode"

	"authservice/constant"
)

type AuthUser struct {
	BearerToken  string `json:"authentication,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`

	User *User `json:"user,omitempty"`
}

type LoginUser struct {
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Password  string `json:"password,omitempty"`
	LoginType string `json:"type"`
	Nonce     string `json:"nonce"`
	OTP       string `json:"otp"`
}

func (l *LoginUser) IsPasswordValid() bool {
	return isPasswordValid(l.Password)
}

func (l *LoginUser) IsPhoneValid() bool {
	return isPhoneValid(l.Phone)
}

func (l *LoginUser) IsEmailValid() bool {
	return isEmailValid(l.Email)
}

type Token struct {
	RefreshToken string `json:"refreshToken,omitempty"`
}

type RefreshMeta struct {
	UserClaims map[string]interface{}
	Expiry     int64
}

func isPhoneValid(phone string) bool {
	length := len(phone)
	if length < 11 || length > 15 {
		return false
	}

	startChar := rune(phone[0])
	if startChar != '+' {
		return false
	}

	for _, digit := range phone[1:] {
		if !unicode.IsDigit(digit) {
			return false
		}
	}

	return true
}

func isEmailValid(email string) bool {
	if len(email) < 8 || !strings.Contains(email, "@") {
		return false
	}

	return true
}

func isPasswordValid(password string) bool {
	if len(password) < 8 {
		return false
	}

	return true
}

func (u *User) isConfirmPasswordValid() bool {
	return u.IsPasswordValid() && u.GetPassword() == u.ConfirmPassword
}

func (u *User) GetRegistrationType() string {
	if u.IsEmailValid() && u.IsPasswordValid() {
		return constant.LoginEmail
	} else if u.IsPhoneValid() && u.IsPasswordValid() {
		return constant.LoginPhone
	} else if u.IsPhoneValid() {
		return constant.LoginPhoneOTP
	} else {
		return constant.LoginInvalid
	}
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
	Nonce           string `json:"nonce"`
}

func (c ChangePasswordRequest) Validate() error {
	if !isPasswordValid(c.CurrentPassword) {
		return fmt.Errorf("current password is not valid")
	}

	if c.CurrentPassword == c.NewPassword {
		return fmt.Errorf("old and new passwords cannot be same")
	}

	if !isPasswordValid(c.NewPassword) {
		return fmt.Errorf("new password is not valid")
	}

	return nil
}

func (c ChangePasswordRequest) ValidateConfirmPassword() error {
	if c.ConfirmPassword != c.NewPassword {
		return fmt.Errorf("passwords do not match")
	}

	if !isPasswordValid(c.NewPassword) {
		return fmt.Errorf("new password is not valid")
	}

	return nil
}