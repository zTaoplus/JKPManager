package common

import (
	"encoding/json"
	"log"

	"zjuici.com/tablegpt/jkpmanager/src/models"
)

func InfoRequestResult(result []byte, idleCount *int) bool {

	var egReplyMessage models.EgWSReplyMessage

	err := json.Unmarshal(result, &egReplyMessage)
	if err != nil {
		panic("cannot unmarshal egReplyMessage")
	}

	content := &models.EgWsReplyStatus{}

	contentBytes, err := json.Marshal(egReplyMessage.Content)

	if err != nil {
		log.Println("Error marshaling content:", err)
	}

	if err := json.Unmarshal(contentBytes, content); err != nil {
		log.Println("Error unmarshaling content:", err)

	}

	// log.Println("Execution State:", content.ExecutionState)
	if content.ExecutionState == "idle" {
		*idleCount++
		if *idleCount == 2 {
			return true
		}
	}
	return false
}

func ExecuteResult(result []byte) bool {
	isDone := false
	var egReplyMessage models.EgWSReplyMessage

	err := json.Unmarshal(result, &egReplyMessage)
	if err != nil {
		panic("cannot unmarshal egReplyMessage")
	}

	if egReplyMessage.ParentHeader.MsgType != "kernel_info_request" {
		return false
	}
	var content interface{}

	switch egReplyMessage.MsgType {
	case "status":
		content = &models.EgWsReplyStatus{}
	case "execute_input":
		content = &models.EgWsReplyExecuteInput{}
	case "execute_reply":
		content = &models.EgWsReplyExecuteReply{}
	case "execute_result":
		content = &models.EgWsReplyExecuteResult{}

	default:
		log.Println("Unknown message type")
	}

	contentBytes, err := json.Marshal(egReplyMessage.Content)
	if err != nil {
		log.Println("Error marshaling content:", err)
	}

	if err := json.Unmarshal(contentBytes, content); err != nil {
		log.Println("Error unmarshaling content:", err)

	}

	// 输出结果
	switch c := content.(type) {
	case *models.EgWsReplyStatus:
		log.Println("Execution State:", c.ExecutionState)
		if c.ExecutionState == "idle" {
			isDone = true
			break
		}
	case *models.EgWsReplyExecuteInput:
		log.Println("Code:", c.Code)
		log.Println("Execution Count:", c.ExecutionCount)
	case *models.EgWsReplyExecuteReply:
		log.Println("Status:", c.Status)
		log.Println("Execution Count:", c.ExecutionCount)
	case *models.EgWsReplyExecuteResult:
		log.Println("Data:", c.Data)
		log.Println("Metadata:", c.Metadata)
		log.Println("Execution Count:", c.ExecutionCount)
	default:
		log.Println("Unknown content type")
	}

	return isDone
}
