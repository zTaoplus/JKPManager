package controllers

import (
	"encoding/json"
	"fmt"
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
		// TODO: after pop , activate with http get

		var kernel models.KernelInfo
		err = json.Unmarshal([]byte(kernelInfo), &kernel)
		if err != nil {
			fmt.Println("Failed to unmarshal kernel JSON:", err)
			return
		}
		log.Println("Try to activate kernel:", kernel.ID)

		httpClient := common.NewHTTPClient(cfg.EGEndpoint)
		_, err = httpClient.Get("/api/kernels/" + kernel.ID)
		if err != nil {
			log.Println("Cannot get the kernel info from Eg: ", err)
		}

		log.Printf("Pre activate kernel %v done", kernel.ID)

		taskClient.StartKernels(1)

		w.Write([]byte(kernelInfo))
	}
}
