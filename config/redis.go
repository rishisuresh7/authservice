package config

type redisConfig struct {
	Port       int
	portString string
	Host       string
	Username   string
	Database   int
	dbString   string
}

func newRedisConfig(missing *[]string) *redisConfig {
	portString, found := getEnv("REDIS_PORT")
	if !found {
		*missing = append(*missing, "REDIS_PORT")
	}

	username, found := getEnv("REDIS_USERNAME")
	if !found {
		*missing = append(*missing, "REDIS_USERNAME")
	}

	host, found := getEnv("REDIS_HOST")
	if !found {
		*missing = append(*missing, "REDIS_HOST")
	}

	dbString, found := getEnv("REDIS_DATABASE_ID")
	if !found {
		dbString = "0"
	}

	return &redisConfig{
		Host:       host,
		Username:   username,
		portString: portString,
		dbString:   dbString,
	}
}
