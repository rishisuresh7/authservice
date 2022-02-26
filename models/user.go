package models

type User struct {
	Id        *string `json:"id" db:"id"`
	Email     *string `json:"email,omitempty" db:"email"`
	Phone     *string `json:"phone,omitempty" db:"phone"`
	LastName  *string `json:"lastName,omitempty" db:"last_name"`
	FirstName *string `json:"firstName,omitempty" db:"first_name"`
}
