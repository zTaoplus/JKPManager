package controllers

import (
	"log"
	"net/http"
	"time"

	"zjuici.com/tablegpt/jkpmanager/src/common"
	"zjuici.com/tablegpt/jkpmanager/src/models"
	"zjuici.com/tablegpt/jkpmanager/src/storage"
)

func PopKernelHandler(cfg *models.Config, taskClient *common.TaskClient, redisClient *storage.RedisClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		poppedKernels, err := redisClient.BRPop(10*time.Second, cfg.RedisKey)

		if err != nil {
			log.Printf("Cannot pop the kernel from redis. error %v", err)
			http.Error(w, "Cannot pop the kernel", http.StatusInternalServerError)
			return
		}
		kernelInfo := poppedKernels[1]

		log.Println("poppedKernels:", kernelInfo)
		taskClient.StartKernels(1)

		w.Write([]byte(kernelInfo))
	}
}
