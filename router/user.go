package router

import (
	"github.com/sirupsen/logrus"

	"authservice/constant"
	"authservice/factory"
	"authservice/handler"
)

func (r *router) userRoutes(f factory.Factory, l*logrus.Logger) {
	r.HandleFunc("/user/login", handler.RegisterUser(f, l)).Methods(constant.POST)
	r.HandleFunc("/user/login/verify", handler.LoginUser(f, l)).Methods(constant.POST)
	r.HandleFunc("/user/logout", handler.LogoutUser(f, l, constant.User)).Methods(constant.GET)
	r.HandleFunc("/user/refresh", handler.RefreshToken(f, l)).Methods(constant.POST)

	// OAuth routes
	handler.InitProviders(f)
	r.HandleFunc("/user/{provider}/auth", handler.OAuth()).Methods(constant.GET)
	r.HandleFunc("/user/{provider}/logout", handler.Logout(l)).Methods(constant.GET)
	r.HandleFunc("/custuseromer/callback", handler.OAuthCallback(f, l)).Methods(constant.GET)
}