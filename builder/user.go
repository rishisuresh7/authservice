package builder

import (
	"fmt"
	"strings"
	"time"
)

type UserBuilder interface {
	Register(user map[string]interface{}) string
	GetUser() string
	Login(loginType string) string
	OAuthRegister() string
	ResetPassword() string
	ChangePassword() string
	UpdateUser(user map[string]interface{}) string
}

type user struct{}

func NewUserBuilder() UserBuilder {
	return &user{}
}

func (u *user) GetUser() string {
	return `SELECT id, first_name, last_name, dob, gender, email, phone, null as address, t_and_c, fb_email, created_at, updated_at,
			verified, pan, aadhar
				FROM users
			WHERE id = $1 OR email = $2 OR phone = $3 LIMIT 1`
}

func (u *user) Register(user map[string]interface{}) string {
	return fmt.Sprintf(`INSERT INTO users(id, first_name, last_name, dob, gender, aadhar, pan, email, fb_email, phone,
		password, t_and_c) VALUES('%s', '%s', '%s', $1, '%s', '%s', '%s', '%s', '%s', '%s', $2, %t)`,
		user["id"], user["firstName"], user["lastName"], user["gender"], user["aadhar"], user["pan"], user["email"],
		user["fbEmail"], user["phone"], user["tAndC"])
}

func (u *user) Login(loginType string) string {
	passwordQuery := "AND password = $3"
	if loginType == "otp" {
		passwordQuery = "OR password = $3" // this is to skip pg from complaining about additional parameter
	}
	return fmt.Sprintf(`SELECT id, first_name, last_name, dob, gender, email, phone, null as address, t_and_c, fb_email, created_at, updated_at,
			verified, pan, aadhar
				FROM users
			WHERE (email = $1 OR phone = $2) %s`, passwordQuery)
}

func (u *user) OAuthRegister() string {
	return ``
}

func (u *user) ResetPassword() string {
	return `UPDATE users SET password = $1 WHERE phone = $2`
}

func (u *user) ChangePassword() string {
	return `UPDATE users SET password = $1 WHERE password = $2 AND id = $3`
}

func (u *user) UpdateUser(user map[string]interface{}) string {
	var updates []string
	for key, value := range user {
		if value != nil {
			switch key {
			case "firstName":
				key = "first_name"
			case "lastName":
				key = "last_name"
			case "fbEmail":
				key = "fb_email"
			case "defaultAddress":
				key = "default_address"
			default:
				continue
			}

			updates = append(updates, fmt.Sprintf("%s = %s", key, getValue(value)))
		}
	}

	return fmt.Sprintf("UPDATE users SET %s, updated_at = '%s' WHERE id = $1", strings.Join(updates, ", "), time.Now().Format(time.RFC3339))
}