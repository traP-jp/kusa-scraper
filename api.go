package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"unicode"

	"github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
)

func getMessages(bot *traqwsbot.Bot) ([]traq.Message, error) {
	var messages []traq.Message
	var before = time.Now()
	for {
		t1 := time.Now()
		res, r, err := bot.API().MessageApi.SearchMessages(context.Background()).Limit(int32(100)).Offset(int32(0)).Before(before).Execute()
		if os.Getenv("DEV") == "true" {

			res, r, err = bot.API().MessageApi.SearchMessages(context.Background()).In("d241f853-302b-49be-a921-cb8855437d62").Limit(int32(100)).Offset(int32(0)).Before(before).Execute()
		}
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
		if !before.After(time.Now().Add(-time.Hour * 14 * 24)) {
			break
		}
	}

	return messages, nil
}

func getYomigana(message string) (string, error) {
	fmt.Println("message: ", message)
	removeStampMessage(&message)
	originalStr := ""
	for _, r := range message {
		if unicode.In(r, unicode.Latin) || unicode.In(r, unicode.Hiragana) || unicode.In(r, unicode.Katakana) || unicode.In(r, unicode.Han) || unicode.In(r, unicode.Number) {
			originalStr += string(r)
		}
	}
	fmt.Println("originalStr: ", originalStr)
	hiraganaRequest := RubyRequest{
		Id:      "kusa",
		Jsonrpc: "2.0",
		Method:  "jlp.furiganaservice.furigana",
		Params: RubyParams{
			Q: originalStr,
		},
	}

	request, err := json.Marshal(hiraganaRequest)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(
		"POST",
		"https://jlp.yahooapis.jp/FuriganaService/V2/furigana",
		bytes.NewBuffer([]byte(request)),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Yahoo AppID: "+os.Getenv("YAHOO_APP_ID"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	responseDataStr, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseData := RubyResponse{}
	err = json.Unmarshal(responseDataStr, &responseData)
	if err != nil {
		return "", err
	}

	finalFurigana := ""
	for _, v := range responseData.Result.Word {
		if v.Furigana == "" {
			finalFurigana += v.Surface
		} else {
			finalFurigana += v.Furigana
		}
	}

	fmt.Println("finalFurigana: ", finalFurigana)
	removeIncompatibleChars(&finalFurigana)
	return finalFurigana, nil
}
