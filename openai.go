package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo"
)

func isSensitive(message string) (bool, error) {
	openAIRequestMessage := OpenAIRequestMessage{
		Role:    "user",
		Content: "次の文章には、18歳未満に対して不適切な内容が含まれていますか？trueかfalseのどちらかで答えてください。trueかfalseのどちらか以外の文字列が回答に含まれないようにしてください。\n\n「" + message + "」",
	}

	openAIRequest := OpenAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []OpenAIRequestMessage{
			openAIRequestMessage,
		},
	}

	openAIRequestStr, err := json.Marshal(openAIRequest)
	if err != nil {
		return false, echo.NewHTTPError(http.StatusInternalServerError, "Failed to marshal OpenAI request", err)
	}

	request, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer([]byte(openAIRequestStr)))
	if err != nil {
		return false, echo.NewHTTPError(http.StatusInternalServerError, "Failed to create request for openAI", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return false, echo.NewHTTPError(http.StatusInternalServerError, "Failed to send request to openAI", err)
	}
	defer response.Body.Close()

	responseStr, err := io.ReadAll(response.Body)
	if err != nil {
		return false, echo.NewHTTPError(http.StatusInternalServerError, "Failed to read response from openAI", err)
	}

	responseObj := OpenAIResponse{}
	err = json.Unmarshal(responseStr, &responseObj)
	if err != nil {
		return false, echo.NewHTTPError(http.StatusInternalServerError, "Failed to unmarshal response from openAI", err)
	}

	responseBool, err := strconv.ParseBool(responseObj.Choices[0].Message.Content)
	if err != nil {
		fmt.Println(responseObj.Choices[0].Message.Content, err)
		return false, nil
	}

	return responseBool, nil
}
