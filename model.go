package main

type RubyRequest struct {
	Id      string     `json:"id"`
	Jsonrpc string     `json:"jsonrpc"`
	Method  string     `json:"method"`
	Params  RubyParams `json:"params"`
}
type RubyParams struct {
	Q string `json:"q"`
}
type RubyResponse struct {
	Id      string     `json:"id"`
	Jsonrpc string     `json:"jsonrpc"`
	Result  RubyResult `json:"result"`
}
type RubyResult struct {
	Word []RubyWord `json:"word"`
}
type RubyWord struct {
	Furigana string `json:"furigana"`
	Roman    string `json:"roman"`
	Surface  string `json:"surface"`
}
type Stamps struct {
	TaskId  string `json:"taskId"`
	StampId string `json:"stampId"`
	Count   int    `json:"count"`
}
