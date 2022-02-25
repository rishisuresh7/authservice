package main

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	ft "github.com/x-cray/logrus-prefixed-formatter"

	"authservice/config"
	"authservice/factory"
	"authservice/middleware"
	"authservice/router"
)

var Version = "0.0.0"

func main() {
	l := logrus.New()
	l.Level = logrus.DebugLevel
	l.Formatter = &ft.TextFormatter{
		ForceFormatting: true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}

	conf, err := config.NewConfig()
	if err != nil {
		l.Fatalf("Error in generating config: %s", err)
	}

	f := factory.NewFactory(l, conf)
	l.Infof("Running auth service server version: %s", Version)

	n := negroni.New()
	n.Use(middleware.NewReqResLogger(l))
	n.UseHandler(router.NewCustomRouter(f, l))
	n.Run(fmt.Sprintf(":%d", conf.Port))
}