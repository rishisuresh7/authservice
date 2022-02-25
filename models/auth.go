package models

import (
	"encoding/json"
	"unicode"
)

type AuthUser struct {
	BearerToken  string `json:"authentication,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`

	User        *User        `json:"user,omitempty"`
}

type RegisterUser struct {
	Phone string `json:"phone"`
	OTP   string `json:"otp"`
	Nonce string `json:"nonce"`
}

type Token struct {
	RefreshToken string `json:"refreshToken,omitempty"`
}

type RefreshMeta struct {
	UserClaims map[string]interface{}
	Expiry     int64
}

func (r *RegisterUser) IsPhoneValid() bool {
	length := len(r.Phone)
	if length < 11 || length > 15 {
		return false
	}

	startChar := rune(r.Phone[0])
	if startChar != '+' {
		return false
	}

	for _, digit := range r.Phone[1:] {
		if !unicode.IsDigit(digit) {
			return false
		}
	}

	return true
}

type UserMeta struct {
	UserId       string
	UserType     string
	BearerToken  string
	RefreshToken string
}

func (u *UserMeta) GetBytes() []byte {
	bytes, _ := json.Marshal(u)
	return bytes
}
