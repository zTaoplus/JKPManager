package models

import "time"

type EgWSReplyHeader struct {
	MsgID    string    `json:"msg_id"`
	MsgType  string    `json:"msg_type"`
	Username string    `json:"username"`
	Session  string    `json:"session"`
	Date     time.Time `json:"date"`
	Version  string    `json:"version"`
}

type EgWSReplyParentHeader struct {
	MsgID    string    `json:"msg_id"`
	MsgType  string    `json:"msg_type"`
	Username string    `json:"username"`
	Session  string    `json:"session"`
	Date     time.Time `json:"date"`
	Version  string    `json:"version"`
}

type EgWsReplyStatus struct {
	ExecutionState string `json:"execution_state"`
}

type EgWsReplyExecuteInput struct {
	Code           string `json:"code"`
	ExecutionCount int    `json:"execution_count"`
}

type EgWsReplyExecuteReply struct {
	Status          string                 `json:"status"`
	ExecutionCount  int                    `json:"execution_count"`
	UserExpressions map[string]interface{} `json:"user_expressions"`
	Payload         []interface{}          `json:"payload"`
}

type EgWsReplyExecuteResult struct {
	Data           interface{} `json:"data"`
	Metadata       interface{} `json:"metadata"`
	ExecutionCount int         `json:"execution_count"`
}

type EgWSReplyMessage struct {
	Header       EgWSReplyHeader        `json:"header"`
	MsgID        string                 `json:"msg_id"`
	MsgType      string                 `json:"msg_type"`
	ParentHeader EgWSReplyParentHeader  `json:"parent_header"`
	Metadata     interface{}            `json:"metadata"`
	Content      map[string]interface{} `json:"content"`

	Buffers []interface{} `json:"buffers"`
	Channel string        `json:"channel"`
}

type EgWsSendContent struct {
	Code            string                 `json:"code"`
	Silent          bool                   `json:"silent"`
	StoreHistory    bool                   `json:"store_history"`
	UserExpressions map[string]interface{} `json:"user_expressions"`
	AllowStdin      bool                   `json:"allow_stdin"`
}

type EgWsSendHeader struct {
	MsgID   string `json:"msg_id"`
	MsgType string `json:"msg_type"`
}

type EgWsSendMessage struct {
	Header       EgWsSendHeader    `json:"header"`
	ParentHeader map[string]string `json:"parent_header"`
	Metadata     map[string]string `json:"metadata"`
	Content      EgWsSendContent   `json:"content"`
	Channel      string            `json:"channel"`
}
