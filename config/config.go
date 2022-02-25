package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port          int
	TokenSecret   string
	RefreshSecret string

	PgConfig      *pgConfig
	RedisConfig   *redisConfig
	ProvidersConf []*providerConf
}

func NewConfig() (*Config, error) {
	conf, missing, err := getConfigFromEnv()
	if err != nil {
		return nil, fmt.Errorf("NewConfig: %s", err.Error())
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("NewConfig: missing env argument(s): %s", strings.Join(missing, ", "))
	}

	return conf, nil
}

func getConfigFromEnv() (*Config, []string, error) {
	var missing []string
	portString, found := getEnv("PORT")
	if !found {
		missing = append(missing, "PORT")
	}

	tokenSecret, found := getEnv("TOKEN_SECRET")
	if !found {
		missing = append(missing, "TOKEN_SECRET")
	}

	refreshSecret, found := getEnv("REFRESH_SECRET")
	if !found {
		missing = append(missing, "REFRESH_SECRET")
	}

	pgConfig := newPostgresConfig(&missing)
	redisConfig := newRedisConfig(&missing)
	googleProvider := newGoogleAuthProvider(&missing)

	if len(missing) > 0 {
		return nil, missing, nil
	}

	port, err := strconv.Atoi(portString)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value provided for PORT")
	}

	pgPort, err := strconv.Atoi(pgConfig.portString)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value provided for PG_PORT")
	}

	redisPort, err := strconv.Atoi(redisConfig.portString)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value provided for REDIS_PORT")
	}

	redisDatabase, err := strconv.Atoi(redisConfig.dbString)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value provided for REDIS_DATABASE_ID")
	}

	pgConfig.Port = pgPort
	redisConfig.Port = redisPort
	redisConfig.Database = redisDatabase

	return &Config{
		Port:          port,
		TokenSecret:   tokenSecret,
		RefreshSecret: refreshSecret,
		PgConfig:      pgConfig,
		RedisConfig:   redisConfig,
		ProvidersConf: []*providerConf{
			googleProvider,
		},
	}, nil, nil
}

func getEnv(key string) (string, bool) {
	value := os.Getenv(key)
	if value == "" {
		return "", false
	}

	return value, true
}
