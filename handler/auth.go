package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"authservice/factory"
	"authservice/models"
	"authservice/response"
)

func LoginUser(f factory.Factory, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.LoginUser
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			l.Errorf("LoginUser: invalid request payload")
			response.Error{Error: "invalid request"}.ClientError(w)
			return
		}

		if !(user.IsEmailValid() || user.IsPhoneValid()) {
			l.Errorf("LoginUser: payload should not have empty values")
			response.Error{Error: "invalid request"}.ClientError(w)
			return
		}

		user.LoginType = strings.ToLower(user.LoginType)
		if user.LoginType != "otp"{
			if !user.IsPasswordValid() {
				l.Errorf("LoginUser: payload should not have empty values")
				response.Error{Error: "invalid request"}.ClientError(w)
				return
			}
		}

		us := f.User()
		var res interface{}
		if user.LoginType == "otp" {
			res, err = us.LoginWithOTP(r.Context(), &user)
		} else {
			res, err = us.Login(r.Context(), &user)
		}

		if err != nil {
			l.Errorf("LoginUser: unable to login user: %s", err)
			response.Error{Error: "unexpected error happened"}.ServerError(w)
			return
		}

		response.Success{Success: res}.Send(w)
	}
}

func VerifyOTP(f factory.Factory, l *logrus.Logger, isLogin bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.LoginUser
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			l.Errorf("VerifyOTP: invalid request payload: %s", err)
			response.Error{Error: "invalid request"}.ClientError(w)
			return
		}

		usr := f.User()
		valid, err := usr.VerifyOTP(r.Context(), &user)
		if err != nil  {
			l.Errorf("VerifyOTP: unable to verify OTP: %s", err)
			response.Error{Error: "unexpected error happened"}.ServerError(w)
			return
		}

		if !valid {
			l.Errorf("VerifyOTP: unable to verify OTP: Invalid OTP")
			response.Error{Error: "invalid request"}.ClientError(w)
			return
		}

		var res interface{}
		if isLogin {
			user.LoginType = "otp"
			res, err = usr.Login(r.Context(), &user)
		} else {
			res, err = usr.GetResetSecret(r.Context(), user.Phone)
		}

		if err != nil {
			l.Errorf("VerifyOTP: unable to get user/secret: %s", err)
			response.Error{Error: "unexpected error happened"}.ServerError(w)
			return
		}

		response.Success{Success: res}.Send(w)
	}
}

func RegisterUser(f factory.Factory, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			l.Errorf("RegisterUser: invalid request payload: %s", err)
			response.Error{Error: "invalid request"}.ClientError(w)
			return
		}

		err = user.Validate()
		if err != nil {
			l.Errorf("RegisterUser: invalid request payload: %s", err)
			response.Error{Error: "invalid request"}.ClientError(w)
			return
		}

		us := f.User()
		exists, registeredUser, err := us.GetUser(r.Context(), "", user.GetEmail(), user.GetPhone())
		if err != nil {
			l.Errorf("RegisterUser: unable to register user: %s", err)
			response.Error{Error: "unexpected error happened"}.ServerError(w)
			return
		}

		if exists {
			response.Success{Success: registeredUser}.SendExists(w)
			return
		}

		res, err := us.Register(r.Context(), &user)
		if err != nil {
			l.Errorf("RegisterUser: unable to register user: %s", err)
			response.Error{Error: "unexpected error happened"}.ServerError(w)
			return
		}

		response.Success{Success: res}.Send(w)
	}
}

func VerifyToken(f factory.Factory, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		splits := strings.Split(token, "Bearer ")
		if len(splits) < 2 {
			l.Errorf("VerifyToken: invalid token format")
			response.Error{Error: "invalid token"}.UnAuthorized(w)
			return
		}

		auth := f.Authorizer()
		claims, err := auth.ValidateBearerToken(r.Context(), token)
		if err != nil {
			l.Errorf("VerifyToken: unable to verify token: %s", err)
			response.Error{Error: "invalid token"}.UnAuthorized(w)
			return
		}

		const (
			name = "name"
			id   = "id"
		)
		w.Header().Add(name, fmt.Sprintf("%s", claims[name]))
		w.Header().Add(id, fmt.Sprintf("%s", claims[id]))

		response.Success{Success: "validated successfully"}.Send(w)
	}
}

func ChangePassword(f factory.Factory, l *logrus.Logger, isReset bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cpr models.ChangePasswordRequest
		err := json.NewDecoder(r.Body).Decode(&cpr)
		if err != nil {
			l.Errorf("ChangePassword: invalid request payload: %s", err)
			response.Error{Error: "invalid request"}.ClientError(w)
			return
		}

		user := f.User()
		if isReset {
			err = cpr.ValidateConfirmPassword()
			if err != nil {
				l.Errorf("ChangePassword: invalid request payload: %s", err)
				response.Error{Error: "invalid request"}.ClientError(w)
				return
			}

			if cpr.Nonce == "" {
				l.Errorf("ChangePassword: invalid request payload: nonce not available")
				response.Error{Error: "invalid request"}.ClientError(w)
				return
			}

			err = user.ResetPassword(r.Context(), &cpr)
			if err != nil {
				l.Errorf("ChangePassword: unable to reset password: %s", err)
				response.Error{Error: "unexpected error happened"}.ServerError(w)
				return
			}
		} else {
			err = cpr.Validate()
			if err != nil {
				l.Errorf("ChangePassword: invalid request payload: %s", err)
				response.Error{Error: "invalid request"}.ClientError(w)
				return
			}

			err = user.ChangePassword(r.Context(), fmt.Sprintf("%s", r.Header.Get("userId")), &cpr)
			if err != nil {
				l.Errorf("ChangePassword: unable to change password: %s", err)
				response.Error{Error: "unexpected error happened"}.ServerError(w)
				return
			}
		}

		response.Success{Success: "password changed successfully"}.Send(w)
	}
}

func ResetPassword(f factory.Factory, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resetUser models.LoginUser
		err := json.NewDecoder(r.Body).Decode(&resetUser)
		if err != nil {
			l.Errorf("ResetPassword: invalid request payload")
			response.Error{Error: "invalid request"}.ClientError(w)
			return
		}

		if !(resetUser.IsEmailValid() || resetUser.IsPhoneValid()) {
			l.Errorf("ResetPassword: payload should not have empty values")
			response.Error{Error: "invalid request"}.ClientError(w)
			return
		}

		user := f.User()
		nonce, err := user.LoginWithOTP(r.Context(), &resetUser)
		if err != nil {
			l.Errorf("ResetPassword: unable to reset password: %s", err)
			response.Error{Error: "unexpected error happened"}.ServerError(w)
			return
		}

		response.Success{Success: nonce}.Send(w)
	}
}

func UpdateUser(f factory.Factory, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userId, ok := vars["userId"]
		if !ok {
			l.Errorf("UpdateUser: unable to read 'userId' from path")
			response.Error{Error: "invalid request"}.ClientError(w)
			return
		}

		var usr models.User
		err := json.NewDecoder(r.Body).Decode(&usr)
		if err != nil {
			l.Errorf("UpdateUser: unable to decode request payload: %s", err)
			response.Error{Error: "invalid request payload"}.ClientError(w)
			return
		}

		user := f.User()
		usr.Id = &userId
		res, err := user.UpdateUser(r.Context(), &usr)
		if err != nil {
			l.Errorf("UpdateUser: unable to update user: %s", err)
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

		bearerToken := r.Header.Get("Authorization")
		if bearerToken == "" {
			l.Errorf("RefreshToken: invalid request: bearer token is not present")
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

		token, err := authorizer.GetJWT(ctx, claims, bearerToken, user.RefreshToken)
		if err != nil {
			l.Errorf("RefreshToken: unable to generate bearer token: %s", err)
			response.Error{Error: "unexpected error happened"}.ServerError(w)
			return
		}

		response.Success{Success: token}.Send(w)
	}
}

func LogoutUser(f factory.Factory, l *logrus.Logger, allSessions bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorizer := f.Authorizer()
		bearerToken := r.Header.Get("Authorization")
		splits := strings.Split(bearerToken, "Bearer ")
		if len(splits) < 2 {
			l.Errorf("ValidateToken: invalid token format")
			response.Error{Error: "invalid token"}.UnAuthorized(w)
			return
		}

		claims, err := authorizer.ValidateBearerToken(r.Context(), bearerToken)
		if err != nil {
			l.Errorf("LogoutUser: unable to decode JWT: %s", err)
			response.Error{Error: "unauthorized"}.UnAuthorized(w)
			return
		}

		err = authorizer.InvalidateTokens(r.Context(), claims["id"].(string), bearerToken, allSessions)
		if err != nil {
			l.Errorf("LogoutUser: unable to invalidate tokens: %s", err)
			response.Error{Error: "unexpected error happened"}.ServerError(w)
			return
		}

		response.Success{Success: ""}.SendNoContent(w)
	}
}
