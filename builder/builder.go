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
	return "SELECT register_customer($1, null, null, null, null)"
}

func (u *user) Login(loginKey string) string {
	return fmt.Sprintf(`SELECT DISTINCT id, first_name, last_name, phone, email, profile_url,
		(SELECT stripe_customer_id FROM user_tokens WHERE user_id = c.id AND user_type = 'customer' LIMIT 1) as stripe_customer_id,
		(created_at = updated_at) as new_user FROM customers c WHERE %s = $1 LIMIT 1`, loginKey)
}

func (u *user) OAuthRegister() string {
	return `SELECT register_customer(null, $1, $2, $3, $4) as id, true as email_verified`
}