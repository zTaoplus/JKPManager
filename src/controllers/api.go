package controllers

import (
	"log"
	"net/http"

	"zjuici.com/tablegpt/jkpmanager/src/common"
	"zjuici.com/tablegpt/jkpmanager/src/models"
	"zjuici.com/tablegpt/jkpmanager/src/storage"
)

func PopKernelHandler(cfg *models.Config, taskClient *common.TaskClient, redisClient *storage.RedisClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		poppedKernels, err := redisClient.BRPop(cfg.RedisKey)

		if err != nil {
			log.Printf("Cannot pop the kernel from redis. error %v", err)
			http.Error(w, "Cannot pop the kernel", http.StatusInternalServerError)
			return
		}

		log.Println("poppedKernels:", poppedKernels[1])
		taskClient.StartKernels(1)

		w.Write([]byte(poppedKernels[1]))
	}
}
