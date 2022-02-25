package config

type pgConfig struct {
	portString string
	Port       int
	Host       string
	Username   string
	Password   string
	Database   string
}

func newPostgresConfig(missing *[]string) *pgConfig {
	portString, found := getEnv("PG_PORT")
	if !found {
		*missing = append(*missing, "PG_PORT")
	}

	host, found := getEnv("PG_HOST")
	if !found {
		*missing = append(*missing, "PG_HOST")
	}

	username, found := getEnv("PG_USERNAME")
	if !found {
		*missing = append(*missing, "PG_USERNAME")
	}

	password, found := getEnv("PG_PASSWORD")
	if !found {
		*missing = append(*missing, "PG_PASSWORD")
	}

	database, found := getEnv("PG_DATABASE")
	if !found {
		*missing = append(*missing, "PG_DATABASE")
	}

	return &pgConfig{
		portString: portString,
		Host:       host,
		Username:   username,
		Password:   password,
		Database:   database,
	}
}
