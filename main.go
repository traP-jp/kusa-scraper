package main

import (
	"context"
	"fmt"
	"os"

	traqwsbot "github.com/traPtitech/traq-ws-bot"
)

func main() {
	bot, err := traqwsbot.NewBot(&traqwsbot.Options{
		AccessToken: os.Getenv("TRAQ_BOT_TOKEN"),
	})
	if err != nil {
		panic(err)
	}

	channel, _, _ := bot.API().ChannelApi.GetChannels(context.Background()).Execute()
	fo, _ := NewForest(bot)
	list := channel.GetPublic()
	for _, c := range list {
		if !c.GetArchived() {
			fmt.Println(c.Id, fo.id_to_path[c.Id])
		}
	}
}
