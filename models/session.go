package models

import (
	"encoding/json"
)

// maxActiveTokens specifies the maximum no of devices a user can log in concurrently
const maxActiveTokens = 3

type UserMeta struct {
	UserId        string
	LastLoginTime int64
	ActiveTokens  []ActiveToken
}

type ActiveToken struct {
	BearerToken  string
	RefreshToken string
}

func (u *UserMeta) GetBytes() []byte {
	bytes, _ := json.Marshal(u)
	return bytes
}

func (u *UserMeta) ContainsBearerToken(bearerToken string) bool {
	for _, activeToken := range u.ActiveTokens {
		if activeToken.BearerToken == bearerToken {
			return true
		}
	}

	return false
}

func (u *UserMeta) ContainsRefreshToken(refreshToken string) bool {
	for _, activeToken := range u.ActiveTokens {
		if activeToken.RefreshToken == refreshToken {
			return true
		}
	}

	return false
}

func (u *UserMeta) ClearToken(bearerToken string) {
	var newActiveTokens []ActiveToken
	for _, activeToken := range u.ActiveTokens {
		if activeToken.BearerToken != bearerToken {
			newActiveTokens = append(newActiveTokens, activeToken)
		}
	}

	u.ActiveTokens = newActiveTokens
}

func (u *UserMeta) ClearAllTokens() {
	u.ActiveTokens = []ActiveToken{}
}

func (u *UserMeta) AddToken(bearerToken, refreshToken string) {
	noOfTokens := len(u.ActiveTokens)
	if noOfTokens < maxActiveTokens {
		u.ActiveTokens = append([]ActiveToken{
			{
				BearerToken: bearerToken,
				RefreshToken: refreshToken,
			},
		}, u.ActiveTokens...)
	} else {
		u.ActiveTokens = append([]ActiveToken{
			{
				BearerToken: bearerToken,
				RefreshToken: refreshToken,
			},
		}, u.ActiveTokens[0:noOfTokens-1]...)
	}
}

func (u *UserMeta) ReplaceBearerToken(oldBearerToken, newBearerToken, refreshToken string) bool {
	tokenIndex := -1
	for index, activeToken := range u.ActiveTokens {
		if activeToken.RefreshToken == refreshToken && activeToken.BearerToken == oldBearerToken {
			tokenIndex = index
			break
		}
	}

	if tokenIndex > -1 {
		u.ActiveTokens[tokenIndex].BearerToken = newBearerToken
		return true
	}

	return false
}