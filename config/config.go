package config

// Configuration : this struct contains ENV configuration parameters
type Configuration struct {
	PostgresEndpoint string `required:"true"`

	StravaClientId     string `required:"true"`
	StravaClientSecret string `required:"true"`
	StravaCallbackUrl  string `required:"true"`
}
