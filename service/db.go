package service

import (
	"context"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/ent"
	"github.com/sirupsen/logrus"

	_ "modernc.org/sqlite"
)

var EntClient *ent.Client

func InitDB() {
	client, err := ent.Open("sqlite3", config.GetRuntimeConfig().DbDSN)
	if err != nil {
		logrus.Fatalf("failed opening connection to sqlite: %v", err)
	}
	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		logrus.Fatalf("failed creating schema resources: %v", err)
	}

	EntClient = client
}

func Close() {
	EntClient.Close()
}
