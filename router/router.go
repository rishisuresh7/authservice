package router

import (
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"authservice/constant"
	"authservice/factory"
	"authservice/handler"
)

func NewCustomRouter(f factory.Factory, l *logrus.Logger) *router {
	r := &router{
		Router: mux.NewRouter(),
	}
	r.registerRoutes(f, l)

	return r
}

type router struct {
	*mux.Router
}

func (r *router) registerRoutes(f factory.Factory, l *logrus.Logger) {
	r.HandleFunc("/health", handler.Health).Methods(constant.GET)
	r.userRoutes(f, l)
	r.addressRoutes(f, l)
}
