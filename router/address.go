package router

import (
	"github.com/sirupsen/logrus"

	"authservice/constant"
	"authservice/factory"
	"authservice/handler"
)

func (r *router) addressRoutes(f factory.Factory, l *logrus.Logger) {
	tokenValidator := f.TokenValidator()
	r.HandleFunc("/users/{userId}/addresses", tokenValidator.ValidateToken(handler.GetAddresses(f, l))).Methods(constant.GET)
	r.HandleFunc("/users/{userId}/addresses", tokenValidator.ValidateToken(handler.CreateAddress(f, l))).Methods(constant.POST)
	r.HandleFunc("/users/{userId}/addresses/{addressId}", tokenValidator.ValidateToken(handler.GetAddress(f, l))).Methods(constant.GET)
	r.HandleFunc("/users/{userId}/addresses/{addressId}", tokenValidator.ValidateToken(handler.UpdateAddress(f, l))).Methods(constant.PATCH)
	r.HandleFunc("/users/{userId}/addresses/{addressId}", tokenValidator.ValidateToken(handler.DeleteAddress(f, l))).Methods(constant.DELETE)
}