package models

import (
	"encoding/json"
	"fmt"
	"time"
)

type User struct {
	Id             *string    `json:"id" db:"id"`
	Email          *string    `json:"email" db:"email"`
	FBEmail        *string    `json:"fbEmail,omitempty" db:"fb_email"`
	Phone          *string    `json:"phone,omitempty" db:"phone"`
	LastName       *string    `json:"lastName,omitempty" db:"last_name"`
	FirstName      *string    `json:"firstName,omitempty" db:"first_name"`
	Gender         *string    `json:"gender" db:"gender"`
	Pan            *string    `json:"pan" db:"pan"`
	Aadhar         *string    `json:"aadhar" db:"aadhar"`
	Deleted        *bool      `json:"deleted,omitempty" db:"deleted"`
	Verified       *bool      `json:"verified" db:"verified"`
	TAndC          *bool      `json:"tAndC" db:"t_and_c"`
	DOB            *time.Time `json:"dob" db:"dob"`
	DefaultAddress *string    `json:"defaultAddress,omitempty" db:"default_address"`
	Address        *Address   `json:"address" db:"address"`

	Password        *string `json:"password,omitempty" db:"password"`
	ConfirmPassword string  `json:"confirmPassword,omitempty" db:"-"`

	CreatedAt *time.Time `json:"createdAt,omitempty" db:"created_at"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty" db:"updated_at"`
}

func (u *User) Validate() error {
	if !u.IsEmailValid() {
		return fmt.Errorf("invalid email")
	}

	if !u.IsPhoneValid() {
		return fmt.Errorf("invalid phone")
	}

	if !u.IsPasswordValid() {
		return fmt.Errorf("invalid password")
	}

	if !u.isConfirmPasswordValid() {
		return fmt.Errorf("invalid confirm password")
	}

	if u.GetFirstName() == "" {
		return fmt.Errorf("invalid first name")
	}

	if !u.GetTAndC() {
		return fmt.Errorf("invalid terms and conditions")
	}

	if u.GetPan() == "" {
		return fmt.Errorf("invalid PAN")
	}

	if u.GetAadhar() == "" {
		return fmt.Errorf("invalid Aadhar")
	}

	// if u.GetGender() == "" {
	// 	return fmt.Errorf("invalid gender")
	// }

	// if u.GetDOB() == nil {
	// 	return fmt.Errorf("invalid DOB")
	// }

	return nil
}

func (u *User) GetId() string {
	if u.Id != nil {
		return *u.Id
	}

	return ""
}

func (u *User) GetFirstName() string {
	if u.FirstName != nil {
		return *u.FirstName
	}

	return ""
}

func (u *User) GetGender() string {
	if u.Gender != nil {
		return *u.Gender
	}

	return ""
}

func (u *User) GetPan() string {
	if u.Pan != nil {
		return *u.Pan
	}

	return ""
}

func (u *User) GetAadhar() string {
	if u.Aadhar != nil {
		return *u.Aadhar
	}

	return ""
}

func (u *User) GetTAndC() bool {
	if u.TAndC != nil {
		return *u.TAndC
	}

	return false
}

func (u *User) GetLastName() string {
	if u.LastName != nil {
		return *u.LastName
	}

	return ""
}

func (u *User) GetEmail() string {
	if u.Email != nil {
		return *u.Email
	}

	return ""
}

func (u *User) GetFBEmail() string {
	if u.FBEmail != nil {
		return *u.FBEmail
	}

	return ""
}

func (u *User) GetDOB() interface{} {
	if u.DOB != nil {
		return *u.DOB
	}

	return nil
}

func (u *User) GetPhone() string {
	if u.Phone != nil {
		return *u.Phone
	}

	return ""
}

func (u *User) GetPassword() string {
	if u.Password != nil {
		return *u.Password
	}

	return ""
}

func (u *User) GetMap() map[string]interface{} {
	var res map[string]interface{}
	bytes, err := json.Marshal(u)
	if err != nil {
		return res
	}

	err = json.Unmarshal(bytes, &res)
	if err != nil {
		return res
	}

	return res
}

func (u *User) IsPhoneValid() bool {
	return isPhoneValid(u.GetPhone())
}

func (u *User) IsEmailValid() bool {
	return isEmailValid(u.GetEmail())
}

func (u *User) IsPasswordValid() bool {
	return isPasswordValid(u.GetPassword())
}