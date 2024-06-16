package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
	"golang.org/x/text/width"
)

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
	message, _, err := bot.API().MessageApi.GetMessage(context.Background(), citated[27:]).Execute()
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

func removeIncompatibleChars(message *string) {
	re := regexp.MustCompile(`[^a-zA-Z0-9ぁ-ん０-９ａ-ｚＡ-Ｚー]*`)
	*message = re.ReplaceAllString(*message, "")
	*message = width.Fold.String(*message)
}

func removeStampMessage(message *string) {
	re := regexp.MustCompile(`:[a-z\-_\.]*:`)
	*message = re.ReplaceAllString(*message, "")
}
