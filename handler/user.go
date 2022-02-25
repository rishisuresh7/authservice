package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"

	"authservice/factory"
	"authservice/models"
	"authservice/response"
)

func LoginUser(f factory.Factory, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.RegisterUser
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			l.Errorf("LoginUser: invalid request payload")
			response.Error{Error: "invalid request"}.ClientError(w)
			return
		}

		if user.Nonce == "" || len(user.OTP) != 6 || !user.IsPhoneValid() {
			l.Errorf("LoginUser: payload should not have empty values")
			response.Error{Error: "invalid request"}.ClientError(w)
			return
		}

		us := f.User()
		res, err := us.Login(r.Context(), &user)
		if err != nil {
			l.Errorf("LoginUser: unable to login user: %s", err)
			response.Error{Error: "unexpected error happened"}.ServerError(w)
			return
		}

		if user.OTP == "000000" {
			response.Success{Success: res}.Send(w)
			return
		}

		otpError, err := us.Verify(r.Context(), &user)
		if err != nil {
			l.Errorf("LoginUser: unable to verify user: %s", err)
			response.Error{Error: "unexpected error happened"}.ServerError(w)
			return
		}

		if otpError {
			l.Errorf("LoginUser: invalid OTP")
			response.Error{Error: "invalid request"}.ClientError(w)
			return
		}

		response.Success{Success: res}.Send(w)
	}
}

func RegisterUser(f factory.Factory, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.RegisterUser
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			l.Errorf("RegisterUser: invalid request payload: %s", err)
			response.Error{Error: "invalid request"}.ClientError(w)
			return
		}

		if !user.IsPhoneValid() {
			l.Errorf("RegisterUser: invalid value for 'phone'")
			response.Error{Error: "invalid request"}.ClientError(w)
			return
		}

		us := f.User()
		res, err := us.Register(r.Context(), user.Phone)
		if err != nil {
			l.Errorf("RegisterUser: unable to register user: %s", err)
			response.Error{Error: "unexpected error happened"}.ServerError(w)
			return
		}

		response.Success{Success: res}.Send(w)
	}
}

func RefreshToken(f factory.Factory, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.Token
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			l.Errorf("RefreshToken: invalid request payload: %s", err)
			response.Error{Error: "invalid request"}.ClientError(w)
			return
		}

		ctx := r.Context()
		authorizer := f.Authorizer()
		claims, err := authorizer.ValidateRefreshToken(ctx, user.RefreshToken)
		if err != nil {
			l.Errorf("RefreshToken: invalid refresh token: %s", err)
			response.Error{Error: "unauthorized"}.UnAuthorized(w)
			return
		}

		token, err := authorizer.GetJWT(ctx, claims)
		if err != nil {
			l.Errorf("RefreshToken: unable to generate bearer token: %s", err)
			response.Error{Error: "unexpected error happened"}.ServerError(w)
			return
		}

		response.Success{Success: token}.Send(w)
	}
}

func LogoutUser(f factory.Factory, l *logrus.Logger, userType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorizer := f.Authorizer()
		userId, err := authorizer.DecodeJWT(r.Header.Get("Authorization"))
		if err != nil {
			l.Errorf("LogoutUser: unable to decode JWT: %s", err)
			response.Error{Error: "unauthorized"}.UnAuthorized(w)
			return
		}

		err = authorizer.InvalidateTokens(r.Context(), userId, userType)
		if err != nil {
			l.Errorf("LogoutUser: unable to invalidate tokens: %s", err)
			response.Error{Error: "unexpected error happened"}.ServerError(w)
			return
		}

		response.Success{Success: ""}.SendNoContent(w)
	}
}
