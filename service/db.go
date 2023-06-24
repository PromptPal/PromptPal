package service

import (
	"context"
	"strings"

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
	initAdminFromEnv()
}

func initAdminFromEnv() {
	adminList := config.GetRuntimeConfig().AdminList
	if len(adminList) == 0 {
		return
	}

	var uc []*ent.UserCreate
	for _, admin := range adminList {
		c := EntClient.
			User.
			Create().
			SetEmail(admin).
			SetAddr(strings.ToLower(admin)).
			SetLang("en").
			SetName(admin).
			SetLevel(255)
		uc = append(uc, c)
	}
	err := EntClient.
		User.
		CreateBulk(uc...).
		OnConflict().
		DoNothing().
		Exec(context.Background())

	if err != nil {
		logrus.Errorln("failed creating admin list from env: ", err)
	}

}

func Close() {
	EntClient.Close()
}
