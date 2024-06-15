package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
)

type HiraganaRequest struct {
	AppId      string `json:"app_id"`
	RequestId  string `json:"request_id"`
	Sentence   string `json:"sentence"`
	OutputType string `json:"output_type"`
}

type HiraganaResponse struct {
	RequestId  string `json:"request_id"`
	OutputType string `json:"output_type"`
	Converted  string `json:"converted"`
}

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
		if !before.After(time.Now().Add(-time.Hour * 4 * 24)) {
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

func getYomigana(message string) (string, error) {
	hiraganaRequest := HiraganaRequest{
		AppId:      os.Getenv("GOO_APP_ID"),
		RequestId:  "",
		Sentence:   message,
		OutputType: "hiragana",
	}
	fmt.Println(message)

	request, err := json.Marshal(hiraganaRequest)
	if err != nil {
		return "", err
	}
	response, err := http.Post("https://labs.goo.ne.jp/api/hiragana", "application/json", bytes.NewBuffer([]byte(request)))
	if err != nil {
		fmt.Println("Error sending request")
		return "", err
	}
	defer response.Body.Close()
	responseDataStr, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	responseData := HiraganaResponse{}
	err = json.Unmarshal(responseDataStr, &responseData)
	if err != nil {
		return "", err
	}
	fmt.Println(string(responseDataStr))

	return responseData.Converted, nil
}

//return: citated, image, isNeedToRemove
func processLinkInMessage(message *string) (string, string, bool) {
	re := regexp.MustCompile(`(http|https)://.*`)
	pathMessages := re.FindAllString(*message, -1)
	var citated, image string

	for _, path := range pathMessages {
		re := regexp.MustCompile(`https://q.trap.jp/messages/(.*)`)
		cites := re.FindAllString(path, -1)
		if len(cites) > 1 {
			return "", "", true
		}
		if len(cites) == 1 {
			citated = cites[0]
			re = regexp.MustCompile(`\nhttps://q.trap.jp/messages/(.*)`)
			*message = re.ReplaceAllString(*message, "")
			continue
		}

		re = regexp.MustCompile(`https://q.trap.jp/files/(.*)`)
		images := re.FindAllString(path, -1)
		if len(images) > 1 {
			return "", "", true
		}
		if len(images) == 1 {
			image = images[0]
			re = regexp.MustCompile(`\nhttps://q.trap.jp/files/(.*)`)
			*message = re.ReplaceAllString(*message, "")
			continue
		}
 
		return "", "", true			
	}
	
	return citated, image, false
}