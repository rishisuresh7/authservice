package builder

import (
	"fmt"
	"strings"
)

type AddressBuilder interface {
	GetAddresses() string
	GetAddress() string
	CreateAddress(addr map[string]interface{}) string
	UpdateAddress(addr map[string]interface{}) string
	RemoveDefaults() string
	DeleteAddress() string
}

type address struct {}

func NewAddressBuilder() AddressBuilder {
	return &address{}
}

func (a *address) GetAddresses() string {
	return `SELECT * FROM addresses WHERE user_id = $1 AND deleted = false ORDER BY created_at`
}

func (a *address) GetAddress() string {
	return `SELECT * FROM addresses WHERE user_id = $1 AND id = $2 AND deleted = false`
}

func (a *address) CreateAddress(addr map[string]interface{}) string {
	return fmt.Sprintf(`INSERT INTO addresses(id, user_id, name, phone, first_line, second_line, address_string, city, state, country,
                      pincode, lat, long, is_default) VALUES ($2, $1, '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%f', '%f', %s)`,
					  addr["name"], addr["phone"], addr["firstLine"], addr["secondLine"], addr["fullAddress"], addr["city"], addr["state"], addr["country"],
					  addr["pincode"], addr["latitude"], addr["longitude"], getValue(addr["isDefault"]))
}

func (a *address) RemoveDefaults() string {
	return `UPDATE addresses SET is_default = false WHERE is_default = true AND user_id = $1`
}

func (a *address) UpdateAddress(addr map[string]interface{}) string {
	var updates []string
	for key, value := range addr {
		switch key {
		case "fullAddress":
			key = "address_string"
		case "latitude":
			key = "lat"
		case "longitude":
			key = "long"
		case "firstLine":
			key = "first_line"
		case "secondLine":
			key = "second_line"
		case "updatedAt":
			key = "updated_at"
		case "isDefault":
			key = "is_default"
		}
		updates = append(updates, fmt.Sprintf("%s = %s", key, getValue(value)))
	}

	return fmt.Sprintf("UPDATE addresses SET %s WHERE id = $2 AND user_id = $1", strings.Join(updates, ", "))
}

func (a *address) DeleteAddress() string {
	return `UPDATE addresses SET deleted = true WHERE user_id = $1 AND id = $2 AND is_default = false`
}
