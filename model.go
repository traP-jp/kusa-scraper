package main

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

type Stamps struct {
	TaskId  string `json:"taskId"`
	StampId string `json:"stampId"`
	Count   int    `json:"count"`
}
