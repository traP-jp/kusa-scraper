package main

import (
	"fmt"
	"log"

	"github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
	"github.com/traPtitech/traq-ws-bot/payload"
)

func updateHandrer(bot *traqwsbot.Bot, p *payload.MessageCreated) {
	fmt.Println(p.Message.Text)
	allMessages, _ := getMessages(bot)
	target := "6308a443-69f0-45e5-866f-56cc2c93578f"

	for _, message := range allMessages {
		wcount := 0
		for _, v := range message.Stamps {
			if v.GetStampId() == target {
				wcount++
			}
		}

		fmt.Println(message.Content)
		if wcount >= 15 {
			insertTask(message, bot)
		}
	}
	simplePost(bot, p.Message.ChannelID, "completed")
}
func insertTask(message traq.Message, bot *traqwsbot.Bot) {
	citated, image, isNeedToRemove := processLinkInMessage(&message.Content)
	if isNeedToRemove {
		return
	}
	if isContainsCodeBrocks(message.Content) || isConstainsMdTable(message.Content) {
		return
	}
	var err error
	if citated != "" {
		citated, err = getCitetedMessage(citated, bot)
		if err != nil {
			panic(err)
		}
	}

	processMentionToPlainText(&citated)
	removeTex(&citated)

	yomi, err := getYomigana(message.Content)
	if err != nil {
		panic(err)
	}
	user := usersMap[message.UserId]
	userGrade := gradeMap[message.UserId]
	iconUri := "https://q.trap.jp/api/v3/public/icon/" + user.Name

	count := 0
	err = db.QueryRow("SELECT COUNT(*) FROM tasks WHERE messageId = ?", message.Id).Scan(&count)

	if err != nil {
		log.Fatalf("DB Error: %s", err)
	}
	if count == 0 {
		_, err = db.Exec("INSERT INTO tasks (content, yomi, iconUri, authorDisplayName, grade, authorName, updatedAt, level, isSensitive,citated, image, messageId) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", message.Content, yomi, iconUri, user.DisplayName, userGrade.Name, user.Name, message.UpdatedAt, 1, false, citated, image, message.Id)
		if err != nil {
			panic(err)
		}

		stampsMap := getStampsData(message.Stamps)
		for stampId, count := range stampsMap {
			_, err = db.Exec("INSERT INTO stamps (taskId, stampId, count) VALUES (?, ?, ?)", message.Id, stampId, count)
			if err != nil {
				panic(err)
			}
		}
	}
}
