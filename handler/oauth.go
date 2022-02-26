package handler

import (
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/sirupsen/logrus"

	"authservice/factory"
	"authservice/models"
	"authservice/response"
)

func InitProviders(f factory.Factory) {
	providers := make([]goth.Provider, 0)
	callbackUrl := "http://localhost:9003/user/callback"
	for _, provider := range f.Config().ProvidersConf {
		switch provider.Name {
		case "google":
			pr := google.New(provider.ClientId, provider.ClientSecret, callbackUrl, "email", "profile")
			providers = append(providers, pr)
		}
	}

	goth.UseProviders(providers...)
}

func Logout(l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := gothic.Logout(w, r)
		if err != nil {
			l.Errorf("Logout: unable to logout user: %s", err)
			response.Error{Error: "unexpected error happened"}.ServerError(w)
			return
		}

		response.Success{Success: "logged out successfully"}.Send(w)
	}
}

func OAuth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := time.Now().String()
		maxAge := 1 * 30
		store := sessions.NewCookieStore([]byte(key))
		store.MaxAge(maxAge)
		store.Options.Path = "/"
		store.Options.HttpOnly = true // HttpOnly should always be enabled
		store.Options.Secure = false
		gothic.Store = store
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		gothic.BeginAuthHandler(w, r)
	}
}

func OAuthCallback(f factory.Factory, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gothUser, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			l.Errorf("OAuthCallback: unable to authenticate user: %s", err)
			response.Error{Error: "unauthorized"}.UnAuthorized(w)
			return
		}

		user := models.User{
			Email:      &gothUser.Email,
			LastName:   &gothUser.LastName,
			FirstName:  &gothUser.FirstName,
		}

		us := f.User()
		authUser, err := us.OAuthLogin(r.Context(), user)
		if err != nil {
			l.Errorf("OAuthCallback: unable to save user info: %s", err)
			response.Error{Error: "unexpected error happened"}.ServerError(w)
			return
		}

		response.Success{Success: authUser}.Send(w)
	}
}
