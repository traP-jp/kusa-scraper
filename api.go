package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
)

func getMessages(bot *traqwsbot.Bot) ([]traq.Message, error) {
	var messages []traq.Message
	var before = time.Now()
	for {
		t1 := time.Now()
		res, r, err := bot.API().MessageApi.SearchMessages(context.Background()).From("2e0c6679-166f-455a-b8b0-35cdfd257256").Limit(int32(100)).Offset(int32(0)).Before(before).Execute()

		// res, r, err := bot.API().MessageApi.SearchMessages(context.Background()).Limit(int32(100)).Offset(int32(0)).Before(before).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `ChannelApi.GetMessages``: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		}

		fmt.Println(time.Since(t1))
		if err != nil {
			return nil, err
		}
		if len(res.Hits) == 0 {
			break
		}
		messages = append(messages, res.Hits...)
		time.Sleep(time.Millisecond * 100)
		before = messages[len(messages)-1].CreatedAt
		fmt.Println(len(messages))
		if !before.After(time.Now().Add(-time.Hour * 28 * 24)) {
			break
		}
	}

	return messages, nil
}
func simplePost(bot *traqwsbot.Bot, c string, s string) (x string) {
	q, r, err := bot.API().
		MessageApi.
		PostMessage(context.Background(), c).
		PostMessageRequest(traq.PostMessageRequest{
			Content: s,
		}).
		Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	return q.Id
}
