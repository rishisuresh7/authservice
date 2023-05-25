package factory

import (
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-redis/redis/v8"
	pg "github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"

	"authservice/address"
	"authservice/auth"
	"authservice/builder"
	"authservice/config"
	"authservice/helper"
	"authservice/middleware"
	"authservice/repository"
	"authservice/user"
)

type Factory interface {
	Config() *config.Config
	PostgresQueryer() repository.PostgresQueryer
	RedisQueryer() repository.RedisQueryer
	User() user.User
	Address() address.Address
	Helper() helper.Helper
	Authorizer() auth.Authorizer
	TokenValidator() *middleware.TokenValidator
}

type factory struct {
	logger *logrus.Logger

	pgConn     *pg.Pool
	awsSession *session.Session
	redisConn  *redis.Client
	config     *config.Config
}

func NewFactory(l *logrus.Logger, conf *config.Config) Factory {
	return &factory{
		logger: l,
		config: conf,
	}
}

func (f *factory) Config() *config.Config {
	return f.config
}

func (f *factory) PostgresQueryer() repository.PostgresQueryer {
	d, err := f.pgDriver()
	if err != nil {
		log.Fatalf("Unable to establish connection to postgres: %s", err)
	}

	return repository.NewPgQueryer(d)
}

func (f *factory) RedisQueryer() repository.RedisQueryer {
	d, err := f.redisDriver()
	if err != nil {
		log.Fatalf("Unable to establish connection to redis: %s", err)
	}

	return repository.NewRedisQueryer(d)
}

func (f *factory) User() user.User {
	return user.NewUser(builder.NewUserBuilder(), f.PostgresQueryer(), f.RedisQueryer(), f.Helper())
}

func (f *factory) Address() address.Address {
	return address.NewAddress(builder.NewAddressBuilder(), f.Helper(), f.PostgresQueryer())
}

func (f *factory) Helper() helper.Helper {
	return helper.NewHelper(f.logger, f.RedisQueryer(), f.config.TokenSecret, f.config.RefreshSecret)
}

func (f *factory) Authorizer() auth.Authorizer {
	return auth.NewAuthorizer(f.Helper(), f.RedisQueryer())
}

func (f *factory) TokenValidator() *middleware.TokenValidator {
	return middleware.NewTokenValidator(f.logger, f.Authorizer())
}