package config

// Config : this struct contains ENV configuration parameters
type Config struct {
	DeploymentType string `required:"true" default:"production"`

	PostgresHost       string
	PostgresUser       string
	PostgresPassword   string
	PostgresPort       int64
	PostgresDb         string
	PostgresRequireSSL string `default:"require"`

	StravaClientID     string
	StravaClientSecret string
	CallbackURL        string
	StravaWebhookURL   string
}
