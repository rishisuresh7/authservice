package config

type providerConf struct {
	Name         string
	ClientId     string
	ClientSecret string
}

func newGoogleAuthProvider(missing *[]string) *providerConf {
	clientId, found := getEnv("GOOGLE_CLIENT_ID")
	if !found {
		*missing = append(*missing, "GOOGLE_CLIENT_ID")
	}

	clientSecret, found := getEnv("GOOGLE_CLIENT_SECRET")
	if !found {
		*missing = append(*missing, "GOOGLE_CLIENT_SECRET")
	}

	return &providerConf{
		Name:         "google",
		ClientId:     clientId,
		ClientSecret: clientSecret,
	}
}
