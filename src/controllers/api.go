package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"zjuici.com/tablegpt/jkpmanager/src/common"
	"zjuici.com/tablegpt/jkpmanager/src/models"
	"zjuici.com/tablegpt/jkpmanager/src/storage"
)

func PopKernelHandler(cfg *models.Config, httpClient *common.HTTPClient, redisClient *storage.RedisClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		var owner models.OwnerUser
		err := json.NewDecoder(r.Body).Decode(&owner)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("Get user id: %v", owner.ID)

		var poppedKernel string

		// TODO: break if executing times more than xxx
		for {
			poppedKernel, err = redisClient.RPop(cfg.RedisKey)
			if err != nil {
				timeDuration := 300 * time.Millisecond
				log.Printf("Cannot pop the kernel from redis. waiting %v", timeDuration)
				time.Sleep(timeDuration)
				continue

			}
			break
		}

		// create new kernel
		go common.StartKernels(cfg, httpClient, redisClient, 1)

		var kernel models.KernelInfo

		err = json.Unmarshal([]byte(poppedKernel), &kernel)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		// TODO:  get this value from env
		workingDirCmdCode := "%mkdir -p " + owner.ID + "\n%cd " + owner.ID
		workingDirWsMessage := models.EgWsSendMessage{
			Header: models.EgWsSendHeader{
				MsgID:   strings.ReplaceAll(uuid.New().String(), "-", ""),
				MsgType: "execute_request",
			},
			ParentHeader: make(map[string]string),
			Metadata:     make(map[string]string),
			Content: models.EgWsSendContent{
				Code:            workingDirCmdCode,
				Silent:          false,
				StoreHistory:    false,
				UserExpressions: make(map[string]interface{}),
				AllowStdin:      false,
			},
			Channel: "shell",
		}

		wsClient := common.NewWebSocketClient(cfg.EGWSEndpoint + "/api/kernels/" + kernel.ID + "/channels")

		defer wsClient.Close()

		err = wsClient.Connect()

		// TODO: if connect ws is err, remove this kernel id and re pop new kernel record from redis?

		if err != nil {
			log.Printf("Cannot connect to the websocket: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return

		}

		msgBody, err := json.Marshal(workingDirWsMessage)
		if err != nil {
			log.Printf("Cannot marshal workingDirWsMessage: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// TODO: if send is err, also execute logic like connect ws err

		err = wsClient.Send(msgBody)
		if err != nil {
			panic("cannot send workingDirWsMessage")
		}

		// handler info request message？
		for {
			select {
			case message := <-wsClient.ResultChan:
				if common.ExecuteResult(message) {
					jsonData, err := json.Marshal(kernel)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					w.Write(jsonData)
					return
				}
			case <-time.After(3 * time.Second):
				log.Printf("Waiting Timeout")
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

	}
}
