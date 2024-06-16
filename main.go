package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
	"github.com/traPtitech/traq-ws-bot/payload"
)

var (
	db       *sqlx.DB
	usersMap = make(map[string]traq.User)
	gradeMap = make(map[string]traq.UserGroup)
)

func main() {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Fatal(err)
	}

	conf := mysql.Config{
		User:                 os.Getenv("NS_MARIADB_USER"),
		Passwd:               os.Getenv("NS_MARIADB_PASSWORD"),
		Net:                  "tcp",
		Addr:                 os.Getenv("NS_MARIADB_HOSTNAME") + ":" + os.Getenv("NS_MARIADB_PORT"),
		DBName:               os.Getenv("NS_MARIADB_DATABASE"),
		ParseTime:            true,
		Collation:            "utf8mb4_unicode_ci",
		Loc:                  jst,
		AllowNativePasswords: true,
	}

	_db, err := sqlx.Open("mysql", conf.FormatDSN())

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("conntected")
	db = _db
	bot, err := traqwsbot.NewBot(&traqwsbot.Options{
		AccessToken: os.Getenv("TRAQ_BOT_TOKEN"),
	})
	if err != nil {
		panic(err)
	}
	bot.OnMessageCreated(func(p *payload.MessageCreated) {
		fmt.Println(p.Message.Text)
		cmd := strings.Split(p.Message.Text, " ")
		if cmd[1] == "update" {
			updateHandrer(bot)
		}
	})

	users, resp, err := bot.API().UserApi.GetUsers(context.Background()).Execute()
	if err != nil {
		fmt.Println(err)
		fmt.Println("HTTP Response: ", resp)
	}
	for _, user := range users {
		usersMap[user.Id] = user
	}

	groups, resp, err := bot.API().GroupApi.GetUserGroups(context.Background()).Execute()
	if err != nil {
		fmt.Println(err)
		fmt.Println("HTTP Response: ", resp)
	}
	gradeGroups := []traq.UserGroup{}
	for _, group := range groups {
		if group.Type == "grade" {
			gradeGroups = append(gradeGroups, group)
		}
	}
	for _, group := range gradeGroups {
		for _, member := range group.Members {
			gradeMap[member.Id] = group
		}
	}
	if os.Getenv("DEV") == "true" {
		updateHandrer(bot)
	}
	if err := bot.Start(); err != nil {
		panic(err)
	}
}
