package builder

import "fmt"

type UserBuilder interface {
	Register() string
	Login(loginKey string) string
	OAuthRegister() string
}

type user struct{}

func NewUserBuilder() UserBuilder {
	return &user{}
}

func (u *user) Register() string {
	return `INSERT INTO users(id, phone) VALUES($2, $1)`
}

func (u *user) Login(loginKey string) string {
	return fmt.Sprintf(`SELECT DISTINCT id, first_name, last_name, phone FROM users WHERE %s = $1 LIMIT 1`, loginKey)
}

func (u *user) OAuthRegister() string {
	return ``
}
