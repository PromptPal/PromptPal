package config

import (
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type RuntimeConfig struct {
	PublicDomain  string   `envconfig:"PUBLIC_DOMAIN" default:"localhost:7788"`
	DbType        string   `envconfig:"DB_TYPE" required:"true"`
	DbDSN         string   `envconfig:"DB_DSN" required:"true"`
	JwtTokenKey   []byte   `envconfig:"JWT_TOKEN_KEY" required:"true"`
	HashidSalt    string   `envconfig:"HASHID_SALT" required:"true"`
	AdminList     []string `envconfig:"ADMIN_LIST"`
	OpenAIBaseURL string   `envconfig:"OPENAI_BASE_URL" default:"https://api.openai.com/v1"`

	SSOGoogleClientID     string `envconfig:"SSO_GOOGLE_CLIENT_ID"`
	SSOGoogleClientSecret string `envconfig:"SSO_GOOGLE_CLIENT_SECRET"`
	SSOGoogleCallbackURL  string `envconfig:"SSO_GOOGLE_CALLBACK_URL"`
}

var runtimeConfig RuntimeConfig

func GetRuntimeConfig() RuntimeConfig {
	return runtimeConfig
}

func SetupConfig(isTesting bool) {
	var rc RuntimeConfig
	godotenv.Load(".env")
	if isTesting {
		godotenv.Load(".env.testing")
	}
	err := envconfig.Process("pp", &rc)
	if err != nil {
		logrus.Panicln(err)
	}

	if strings.HasPrefix(rc.DbDSN, "sqlite") || strings.HasPrefix(rc.DbType, "sqlite") {
		logrus.Warningln("sqlite3 is not supported anymore, and will be removed in the future. please move to PostgresSQL, for more detail please visit: https://promptpal.github.io/blog/promptPal-1-7")
	}

	runtimeConfig = rc
}
