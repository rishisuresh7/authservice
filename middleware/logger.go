package middleware

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

type requestResponseLogger struct {
	logger *logrus.Logger
}

func NewReqResLogger(l *logrus.Logger) Middleware {
	return &requestResponseLogger{logger: l}
}

func (rrl *requestResponseLogger) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	startTime := time.Now()
	rrl.logger.WithFields(requestFields(r)).Infof("Request")
	
	next(w, r)

	fields := responseFields(r, w.(negroni.ResponseWriter))
	fields["Duration"] = int64(time.Since(startTime) / time.Millisecond)
	rrl.logger.WithFields(fields).Infof("Response")
}

func requestFields(r *http.Request) logrus.Fields {
	fields := logrus.Fields{}
	fields["Client"] = r.RemoteAddr
	fields["Method"] = r.Method
	fields["URL"] = r.URL.String()
	fields["Referrer"] = r.Referer()
	fields["User-Agent"] = r.UserAgent()

	return fields
}

func responseFields(r *http.Request, w negroni.ResponseWriter) logrus.Fields {
	fields := logrus.Fields{}
	fields["Method"] = r.Method
	fields["URL"] = r.URL.String()
	fields["StatusCode"] = w.Status()
	fields["Size"] = w.Size()

	return fields
}
