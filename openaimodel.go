package main

type OpenAIRequest struct {
	Model    string                 `json:"model"`
	Messages []OpenAIRequestMessage `json:"messages"`
}

type OpenAIRequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Id      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []OpenAIResponseChoice `json:"choices"`
}

type OpenAIResponseChoice struct {
	Index   int                           `json:"index"`
	Message []OpenAIResponseChoiceMessage `json:"message"`
}

type OpenAIResponseChoiceMessage struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}
