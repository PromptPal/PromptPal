package service

import (
	"context"
	"strings"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/ent"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

var EntClient *ent.Client

func InitDB() {
	dsn := config.GetRuntimeConfig().DbDSN
	client, err := ent.Open(config.GetRuntimeConfig().DbType, dsn)
	if err != nil {
		logrus.Fatalf("failed opening connection to sqlite: %v", err)
	}
	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		logrus.Fatalf("failed creating schema resources: %v", err)
	}

	EntClient = client
	logrus.Infoln("Connected to database")
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
			SetPhone("").
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
