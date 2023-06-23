package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type RuntimeConfig struct {
	PublicDomain string `envconfig:"PUBLIC_DOMAIN" default:"localhost:7788"`
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
