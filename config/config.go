package config

import (
	"fmt"
	"io/ioutil"
	"log"
)

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

// ReadSecret : Read a file and return it's content as string - used for Docker secrets
func ReadSecret(path string) (result string) {
	data, err := ioutil.ReadFile(fmt.Sprintf("/run/secret/%v", result))
	if err != nil {
		log.Fatalf("Could not read secret: %V", err)
	}
	return string(data)
}
