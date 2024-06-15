package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
	"github.com/traPtitech/traq-ws-bot/payload"
)

func getMessages(bot *traqwsbot.Bot) ([]traq.Message, error) {
	var messages []traq.Message
	var before = time.Now()
	for {
		t1 := time.Now()
		res, r, err := bot.API().MessageApi.SearchMessages(context.Background()).Limit(int32(100)).Offset(int32(0)).Before(before).Execute()

		//res, r, err := bot.API().MessageApi.SearchMessages(context.Background()).Limit(int32(100)).Offset(int32(0)).From("2e0c6679-166f-455a-b8b0-35cdfd257256").Before(before).Execute()
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
		if !before.After(time.Now().Add(-time.Hour * 7 * 24)) {
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

// return: citated, image, isNeedToRemove
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
			re = regexp.MustCompile(`https://q.trap.jp/messages/(.*)`)
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
			re = regexp.MustCompile(`https://q.trap.jp/files/(.*)`)
			*message = re.ReplaceAllString(*message, "")
			continue
		}

		return "", "", true
	}

	return citated, image, false
}

func getStampsData(stamps []traq.MessageStamp) map[string]int {
	stampsData := make(map[string]int)
	for _, stamp := range stamps {
		if _, ok := stampsData[stamp.StampId]; !ok {
			stampsData[stamp.StampId] = 0
		}
		stampsData[stamp.StampId] += int(stamp.Count)
	}
	return stampsData
}

func updateHandrer(bot *traqwsbot.Bot, p *payload.MessageCreated) {
	allMessages, _ := getMessages(bot)
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
	}
	simplePost(bot, p.Message.ChannelID, "completed")
}
