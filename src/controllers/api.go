package controllers

import (
	"context"
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

func PopKernelHandler(cfg *models.Config, taskClient *scheduler.TaskClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		redisClient := storage.GetRedisClient()

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")

		var kernelInfo string

		httpClient := common.NewHTTPClient(cfg.EGEndpoint)

		for {
			poppedKernels, err := redisClient.Client.BRPop(ctx, 10*time.Second, cfg.RedisKey).Result()

			if err != nil {
				log.Printf("Cannot pop the kernel from redis. error %v", err)
				http.Error(w, "Cannot pop the kernel", http.StatusInternalServerError)
				return
			}
			// if popped, start 1 kernels
			log.Println("Popped 1 kernel,start create 1 kernel task")
			taskClient.StartKernels(1)

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

		w.Write([]byte(kernelInfo))
	}
}
