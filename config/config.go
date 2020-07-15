package config

// Configuration : this struct contains ENV configuration parameters
type Config struct {
	PostgresEndpoint   string `required:"true"`
	PostgresUser       string `required:"true"`
	PostgresPassword   string `required:"true"`
	PostgresPort       int    `default:"5432"`
	PostgresDb         string `required:"true"`
	PostgresRequireSSL string `default:"require"`

	StravaClientID     string `required:"true"`
	StravaClientSecret string `required:"true"`
	CallbackURL        string `required:"true"`
	StravaWebhookURL   string `required:"true"`
}
