package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Address struct {
	Id          *string  `json:"id" db:"id"`
	UserId      *string  `json:"-" db:"user_id"`
	Name        *string  `json:"name" db:"name"`
	FirstLine   *string  `json:"firstLine" db:"first_line"`
	SecondLine  *string  `json:"secondLine" db:"second_line"`
	FullAddress *string  `json:"fullAddress" db:"address_string"`
	City        *string  `json:"city" db:"city"`
	State       *string  `json:"state" db:"state"`
	Country     *string  `json:"country" db:"country"`
	PinCode     *string  `json:"pincode" db:"pincode"`
	Phone       *string  `json:"phone" db:"phone"`
	Latitude    *float32 `json:"latitude" db:"lat"`
	Longitude   *float32 `json:"longitude" db:"long"`
	Deleted     *bool    `json:"-" db:"deleted"`
	IsDefault   *bool    `json:"isDefault" db:"is_default"`

	CreatedAt   *time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   *time.Time `json:"updatedAt" db:"updated_at"`
}

func (a *Address) GetFieldMap() map[string]interface{} {
	addressMap := make(map[string]interface{})
	res, _ := json.Marshal(a)
	_ = json.Unmarshal(res, &addressMap)

	deleteKeys := []string{
		"id",
		"createdAt",
		"deleted",
	}
	for key, value := range addressMap {
		if key == "updatedAt" {
			addressMap[key] = time.Now().Format(time.RFC3339)
			continue
		}

		if key == "isDefault" && value == nil {
			addressMap[key] = false
			continue
		}

		if value == nil {
			deleteKeys = append(deleteKeys, key)
		}
	}

	for _, key := range deleteKeys {
		delete(addressMap, key)
	}

	return addressMap
}


func (a *Address) Validate() error {
	var errorKeys []string
	if !a.isNameValid() {
		errorKeys = append(errorKeys, "name")
	}

	if !a.isPhoneValid() {
		errorKeys = append(errorKeys, "phone")
	}

	if !a.isCityValid() {
		errorKeys = append(errorKeys, "'city'")
	}

	if !a.isStateValid() {
		errorKeys = append(errorKeys, "'state'")
	}

	if !a.isCountryValid() {
		errorKeys = append(errorKeys, "'country'")
	}

	if !a.isPostalCodeValid() {
		errorKeys = append(errorKeys, "'postalCode'")
	}

	if !a.isLatValid() {
		errorKeys = append(errorKeys, "'latitude'")
	}

	if !a.isLongValid() {
		errorKeys = append(errorKeys, "'longitude'")
	}

	if len(errorKeys) > 0 {
		return fmt.Errorf("invalid value for key(s): %s", strings.Join(errorKeys, ", "))
	}

	return nil
}

func (a *Address) isNameValid() bool {
	if a.Name != nil {
		if *a.Name != "" {
			return true
		}
	}

	return false
}

func (a *Address) isPhoneValid() bool {
	if a.Phone != nil {
		if *a.Phone != "" {
			return true
		}
	}

	return false
}

func (a *Address) isCityValid() bool {
	if a.City != nil {
		if *a.City != "" {
			return true
		}
	}

	return false
}

func (a *Address) isStateValid() bool {
	if a.State != nil {
		if *a.State != "" {
			return true
		}
	}

	return false
}

func (a *Address) isCountryValid() bool {
	if a.Country != nil {
		if *a.Country != "" {
			return true
		}
	}

	return false
}

func (a *Address) isPostalCodeValid() bool {
	if a.PinCode != nil {
		if *a.PinCode != "" {
			return true
		}
	}

	return false
}

func (a *Address) isLatValid() bool {
	if a.Latitude != nil {
		return true
	}

	return false
}

func (a *Address) isLongValid() bool {
	if a.Longitude != nil {
		return true
	}

	return false
}
