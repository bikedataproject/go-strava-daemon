package config

// Configuration : this struct contains ENV configuration parameters
type Configuration struct {
	PostgresEndpoint string `required:"true"`

	StravaClientID     string `required:"true"`
	StravaClientSecret string `required:"true"`
	StravaCallbackURL  string `required:"true"`
}
