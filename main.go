package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
	"github.com/traPtitech/traq-ws-bot/payload"
)

var (
	db *sqlx.DB
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
			updateHandrer(bot, p)
		}
	})
	// channel, _, _ := bot.API().ChannelApi.GetChannels(context.Background()).Execute()
	// fo, _ := NewForest(bot)
	// list := channel.GetPublic()
	// for _, c := range list {
	// 	if !c.GetArchived() {
	// 		fmt.Println(c.Id, fo.id_to_path[c.Id])
	// 	}
	// }
	if err := bot.Start(); err != nil {
		panic(err)
	}
}
func updateHandrer(bot *traqwsbot.Bot, p *payload.MessageCreated) {
	allMessages, _ := getMessages(bot)
	t := ":w: > 10\n"
	target := "6308a443-69f0-45e5-866f-56cc2c93578f"

	for _, message := range allMessages {
		wcount := 0
		for _, v := range message.Stamps {
			if v.GetStampId() == target {
				wcount++
			}
		}
		if wcount > 10 {

			citated, image, isNeedToRemove := processLinkInMessage(&message.Content)
			if isNeedToRemove {
				continue
			}

			yomi, err := getYomigana(message.Content)
			if err != nil {
				panic(err)
			}
			t += "https://q.trap.jp/messages/" + message.Id + "\n" + yomi + "\n"

			_, err = db.Exec("INSERT INTO tasks (content, yomi, iconUri, authorDisplayName, grade, authorName, updatedAt, level, isSensitive,citated, image) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", message.Content, yomi, "dummy", "dummy", "dummy", "dummy", message.UpdatedAt, 1, false, citated, image)
			if err != nil {
				panic(err)
			}
		}
	}
	simplePost(bot, p.Message.ChannelID, t)
}
