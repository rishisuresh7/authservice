package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"authservice/auth"
	"authservice/response"
)

type TokenValidator struct {
	auth   auth.Authorizer
	logger *logrus.Logger
}

func NewTokenValidator(l *logrus.Logger, a auth.Authorizer) *TokenValidator {
	return &TokenValidator{
		auth: a,
		logger: l,
	}
}

func (t *TokenValidator) ValidateToken(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userId, ok := vars["userId"]
		if !ok {
			t.logger.Errorf("ValidateToken: unable to read 'userId' from path")
			response.Error{Error: "invalid request"}.ClientError(w)
			return
		}

		token := r.Header.Get("Authorization")
		splits := strings.Split(token, "Bearer ")
		if len(splits) < 2 {
			t.logger.Errorf("ValidateToken: invalid token format")
			response.Error{Error: "invalid token"}.UnAuthorized(w)
			return
		}

		claims, err := t.auth.ValidateBearerToken(r.Context(), token)
		if err != nil {
			t.logger.Errorf("ValidateToken: unable to verify token: %s", err)
			response.Error{Error: "invalid token"}.UnAuthorized(w)
			return
		}

		decodedUserId := fmt.Sprintf("%s", claims["id"])
		if userId != decodedUserId {
			t.logger.Errorf("ValidateToken: invalid user id")
			response.Error{Error: "forbidden"}.Forbidden(w)
			return
		}

		r.Header.Add("userId", decodedUserId)
		r.Header.Add("userName", fmt.Sprintf("%s", claims["name"]))
		next(w, r)
	}
}