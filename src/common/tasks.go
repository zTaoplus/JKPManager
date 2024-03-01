package common

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"zjuici.com/tablegpt/jkpmanager/src/models"
	"zjuici.com/tablegpt/jkpmanager/src/storage"
)

// 定时任务 间隔20分钟 激活一次所有连接
func kernelActivator(egWsEndpoint string, kernel *models.KernelInfo, wg *sync.WaitGroup) {

	defer wg.Done()
	wsUrl := egWsEndpoint + "/api/kernels/" + kernel.ID + "/channels"

	wsClient := NewWebSocketClient(wsUrl)
	defer wsClient.Close()

	err := wsClient.Activate()
	if err != nil {
		log.Printf("Cannot connect to the websocket: %v", err)
		return

	}
	idleCount := 0

	for {
		select {
		case message := <-wsClient.ResultChan:

			if InfoRequestResult(message, &idleCount) {
				log.Println("active the kernel done")
				return
			}
		case <-time.After(3 * time.Second):
			log.Printf("Waiting Timeout")
			return
		}
	}

}

// got values from redis
func KernelActivateTask(cfg *models.Config, redisClient *storage.RedisClient) {
	var wg sync.WaitGroup

	kernelsJSON, err := redisClient.LRange(cfg.RedisKey, 0, -1)

	if err != nil {
		log.Printf("Error when LRange redis: %v", err)
		return
	}

	if len(kernelsJSON) < 3 {
		log.Println("go create kernels, len(kernelsJSON) < 3")
	}

	for _, kernelStr := range kernelsJSON {
		var kernel models.KernelInfo
		err := json.Unmarshal([]byte(kernelStr), &kernel)
		if err != nil {
			fmt.Println("Failed to unmarshal kernel JSON:", err)
			continue
		}
		wg.Add(1)
		go kernelActivator(cfg.EGWSEndpoint, &kernel, &wg)
	}

	wg.Wait()
}
