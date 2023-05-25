package router

import (
	"github.com/sirupsen/logrus"

	"authservice/constant"
	"authservice/factory"
	"authservice/handler"
)

func (r *router) userRoutes(f factory.Factory, l *logrus.Logger) {
	tokenValidator := f.TokenValidator()
	r.HandleFunc("/users", handler.RegisterUser(f, l)).Methods(constant.POST)
	r.HandleFunc("/users/auth", handler.LoginUser(f, l)).Methods(constant.POST)
	r.HandleFunc("/users/logout", handler.LogoutUser(f, l, false)).Methods(constant.GET)
	r.HandleFunc("/users/clear", handler.LogoutUser(f, l, true)).Methods(constant.GET)
	r.HandleFunc("/users/refresh", handler.RefreshToken(f, l)).Methods(constant.POST)
	r.HandleFunc("/users/auth/otp", handler.VerifyOTP(f, l, true)).Methods(constant.POST)
	r.HandleFunc("/users/auth/verify", handler.VerifyToken(f, l)).Methods(constant.GET)
	r.HandleFunc("/users/reset", handler.ResetPassword(f, l)).Methods(constant.POST)
	r.HandleFunc("/users/reset/verify", handler.VerifyOTP(f, l, false)).Methods(constant.POST)
	r.HandleFunc("/users/reset/change", handler.ChangePassword(f, l, true)).Methods(constant.POST)
	r.HandleFunc("/users/{userId}/change", tokenValidator.ValidateToken(handler.ChangePassword(f, l, false))).Methods(constant.POST)
	r.HandleFunc("/users/{userId}", tokenValidator.ValidateToken(handler.UpdateUser(f, l))).Methods(constant.PATCH)

	// OAuth routes
	handler.InitProviders(f)
	r.HandleFunc("/user/{provider}/auth", handler.OAuth()).Methods(constant.GET)
	r.HandleFunc("/user/{provider}/logout", handler.Logout(l)).Methods(constant.GET)
	r.HandleFunc("/user/callback", handler.OAuthCallback(f, l)).Methods(constant.GET)
}
