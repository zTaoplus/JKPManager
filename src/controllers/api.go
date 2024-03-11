package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"zjuici.com/tablegpt/jkpmanager/src/common"
	"zjuici.com/tablegpt/jkpmanager/src/models"
	"zjuici.com/tablegpt/jkpmanager/src/scheduler"
	"zjuici.com/tablegpt/jkpmanager/src/storage"
)

func PopKernelHandler(cfg *models.Config, taskClient *scheduler.TaskClient, redisClient *storage.RedisClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		var kernelInfo string

		httpClient := common.NewHTTPClient(cfg.EGEndpoint)
		var needCreateCount int

		for {
			poppedKernels, err := redisClient.BRPop(10*time.Second, cfg.RedisKey)

			needCreateCount++

			if err != nil {
				log.Printf("Cannot pop the kernel from redis. error %v", err)
				http.Error(w, "Cannot pop the kernel", http.StatusInternalServerError)
				return
			}
			kernelInfo = poppedKernels[1]

			log.Println("poppedKernels:", kernelInfo)

			var kernel models.KernelInfo
			err = json.Unmarshal([]byte(kernelInfo), &kernel)
			if err != nil {
				fmt.Println("Failed to unmarshal kernel JSON:", err)
				return
			}
			log.Println("Try to activate kernel:", kernel.ID)

			_, err = httpClient.Get("/api/kernels/" + kernel.ID)

			if err != nil {
				log.Println("Cannot get the kernel info from Eg: ", err)
				continue
			}
			log.Printf("Pre activate kernel %v done", kernel.ID)
			break
		}

		log.Println("needCreateCount:", needCreateCount)
		taskClient.StartKernels(needCreateCount)

		w.Write([]byte(kernelInfo))
	}
}
