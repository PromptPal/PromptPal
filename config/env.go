package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type RuntimeConfig struct {
	PublicDomain  string   `envconfig:"PUBLIC_DOMAIN" default:"localhost:7788"`
	DbDSN         string   `envconfig:"DB_DSN" required:"true"`
	AdminList     []string `envconfig:"ADMIN_LIST"`
	JwtToken      []byte   `envconfig:"JWT_TOKEN"`
	OpenAIBaseURL string   `envconfig:"OPENAI_BASE_URL" default:"https://api.openai.com/v1"`
}

var runtimeConfig RuntimeConfig

func GetRuntimeConfig() RuntimeConfig {
	return runtimeConfig
}

func init() {
	var rc RuntimeConfig
	err := envconfig.Process("pp", &rc)
	if err != nil {
		logrus.Panicln(err)
	}

	runtimeConfig = rc
}
