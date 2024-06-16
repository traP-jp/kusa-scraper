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
	"strings"
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
		//res, r, err := bot.API().MessageApi.SearchMessages(context.Background()).Limit(int32(100)).Offset(int32(0)).Before(before).Execute()

		res, r, err := bot.API().MessageApi.SearchMessages(context.Background()).Limit(int32(100)).Offset(int32(0)).From("2e0c6679-166f-455a-b8b0-35cdfd257256").Before(before).Execute()
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
	hiraganaRequest := RubyRequest{
		Id:      "kusa",
		Jsonrpc: "2.0",
		Method:  "jlp.furiganaservice.furigana",
		Params: RubyParams{
			Q: message,
		},
	}
	fmt.Println(message)

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
	return finalFurigana, nil
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

func getCitetedMessage(citated string, bot *traqwsbot.Bot) (string, error) {
	message, _, err := bot.API().MessageApi.GetMessage(context.Background(), citated).Execute()
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`(http|https)://[a-zA-Z0-9./-_!'()?&:]*`)
	linkRemovedMessage := re.ReplaceAllString(message.Content, "")
	processMentionToPlainText(&linkRemovedMessage)
	removeTex(&linkRemovedMessage)

	return linkRemovedMessage, nil
}

func processMentionToPlainText(message *string) {
	re := regexp.MustCompile(`!{"type":([^!]*)}`)
	mentions := re.FindAllString(*message, -1)
	for _, mention := range mentions {
		fmt.Println(mention)
		re = regexp.MustCompile(`"raw":"(.*)",( *)"id"`)
		mentionRaw := re.FindString(mention)
		mentionRaw = mentionRaw[8 : len(mentionRaw)-1]
		quoteIndex := strings.Index(mentionRaw, "\"")
		mentionRaw = mentionRaw[:quoteIndex]
		*message = strings.Replace(*message, mention, mentionRaw, 1)
	}
}

func removeTex(message *string) {
	re := regexp.MustCompile(`\$(.*)\$`)
	*message = re.ReplaceAllString(*message, "")
}

func isContainsCodeBrocks(message string) bool {
	re := regexp.MustCompile("```+")
	return re.MatchString(message)
}

func isConstainsMdTable(message string) bool {
	re := regexp.MustCompile(`\|(.*)\|`)
	return re.MatchString(message)
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
		if wcount > 10 {

			citated, image, isNeedToRemove := processLinkInMessage(&message.Content)
			if isNeedToRemove {
				continue
			}
			if isContainsCodeBrocks(message.Content) || isConstainsMdTable(message.Content) {
				continue
			}
			citated, err := getCitetedMessage(citated, bot)
			if err != nil {
				panic(err)
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
	}
	simplePost(bot, p.Message.ChannelID, "completed")
}
