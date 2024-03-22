package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"zjuici.com/tablegpt/jkpmanager/src/common"
	"zjuici.com/tablegpt/jkpmanager/src/models"
	"zjuici.com/tablegpt/jkpmanager/src/storage"
)

type CreatingKernelCount struct {
	creatingCount int
	mu            sync.Mutex
}

func NewCreatingKernelCount() *CreatingKernelCount {
	return &CreatingKernelCount{}
}

func (c *CreatingKernelCount) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.creatingCount++

}

func (c *CreatingKernelCount) Decrement() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.creatingCount > 0 {
		c.creatingCount--
	}

}

func (c *CreatingKernelCount) Get() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.creatingCount
}

type TaskClient struct {
	httpClient            *common.HTTPClient
	redisClient           *storage.RedisClient
	cfg                   *models.Config
	creatingKernelCount   *CreatingKernelCount
	toCreateKernelsChan   chan map[string]interface{}
	toActivateKernelsChan chan string
	done                  chan struct{}
}

func NewTaskClient(cfg *models.Config) *TaskClient {
	return &TaskClient{
		toCreateKernelsChan:   make(chan map[string]interface{}, 200),
		toActivateKernelsChan: make(chan string, 200),
		cfg:                   cfg,
		creatingKernelCount:   NewCreatingKernelCount(),
		done:                  make(chan struct{}),
		httpClient:            common.NewHTTPClient(cfg.EGEndpoint),
		redisClient:           storage.GetRedisClient(),
	}
}

func (t *TaskClient) Start() {
	go t.startKernelsLoop()
	go t.activateKernelsLoop()
	go t.checkAndCreateKernelsLoop()
}

func (t *TaskClient) InitKernels() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// check the
	storedKernelsLen, err := t.redisClient.Client.LLen(ctx, t.cfg.RedisKey).Result()

	if err != nil {
		log.Panicf("Cannot get the kernel count from redis: %v", err)
	}

	log.Printf("Existing Pending Kernel Count: %v, Max Pending Kernel Count: %v", storedKernelsLen, t.cfg.MaxPendingKernels)

	needCreateKernelCount := t.cfg.MaxPendingKernels - int(storedKernelsLen)
	if needCreateKernelCount < 0 {
		log.Println("need to delete :", -needCreateKernelCount)
		t.DeleteKernelByCount(-needCreateKernelCount)

		needCreateKernelCount = 0
	}

	// starting kernel create task
	log.Println("needCreateKernelCount:", needCreateKernelCount)
	err = t.StartKernels(needCreateKernelCount)
	if err != nil {
		log.Panicln("Cannot create kernels:", err)
	}

}

func (t *TaskClient) StartKernels(needCreateKernelCount int) error {

	kernelVolumeMounts, err := json.Marshal([]map[string]string{
		{
			"name":      "shared-vol",
			"mountPath": t.cfg.WorkingDir,
		},
	})
	if err != nil {
		log.Println("Cannot marshal the kernelVolumeMounts")
		return err
	}
	kernelVolumes, err := json.Marshal([]map[string]interface{}{
		{"name": "shared-vol",
			"nfs": map[string]string{
				"server": t.cfg.NFSVolumeServer,
				"path":   t.cfg.NFSMountPath,
			},
		},
	})

	if err != nil {
		log.Println("Cannot marshal the kernelVolumes")
		return err
	}

	data := map[string]interface{}{
		"name": "python_kubernetes",
		"env": map[string]string{
			"KERNEL_NAMESPACE":     t.cfg.KernelNamespace,
			"KERNEL_WORKING_DIR":   t.cfg.WorkingDir,
			"KERNEL_VOLUME_MOUNTS": string(kernelVolumeMounts),
			"KERNEL_VOLUMES":       string(kernelVolumes),
			"KERNEL_IMAGE":         t.cfg.KernelImage,
		},
	}

	for i := 0; i < needCreateKernelCount; i++ {
		t.toCreateKernelsChan <- data
	}

	return nil
}

func (t *TaskClient) ActivateKernels() error {
	log.Printf("Start the scheduled task KernelActivator, activate at intervals of %v seconds.", t.cfg.ActivationInterval)

	ticker := time.NewTicker(time.Duration(t.cfg.ActivationInterval) * time.Second)

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		kernelsJSON, err := t.redisClient.Client.LRange(ctx, t.cfg.RedisKey, 0, -1).Result()

		if err != nil {
			log.Printf("Error when LRange redis: %v", err)
			return err
		}

		for _, kernelStr := range kernelsJSON {
			var kernel models.KernelInfo
			err := json.Unmarshal([]byte(kernelStr), &kernel)
			if err != nil {
				fmt.Println("Failed to unmarshal kernel JSON:", err)
				continue
			}
			t.toActivateKernelsChan <- kernel.ID
		}

		return nil
	}

	return nil
}

func (t *TaskClient) startKernelsLoop() {

	for {
		select {
		case <-t.done:
			return
		case data := <-t.toCreateKernelsChan:
			err := t.createKernel(data)
			if err != nil {
				log.Println("cannot create kernel, err:", err)
			}

		}

	}
}

func (t *TaskClient) activateKernelsLoop() {

	for {
		select {
		case <-t.done:
			return
		case kernelId := <-t.toActivateKernelsChan:
			err := t.activateKernel(kernelId)
			if err != nil {
				log.Printf("Cannot activate kernel,ID: %v, err: %v", kernelId, err)
			}
		}

	}
}

func (t *TaskClient) checkAndCreateKernelsLoop() {
	log.Println("[TASK:checkAndCreateKernelsLoop] task started, timer: ", t.cfg.CheckTaskInterval)

	ticker := time.NewTicker(time.Duration(t.cfg.CheckTaskInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-t.done:
			return
		case <-ticker.C:
			t.checkAndCreateKernels()
		}
	}
}

func (t *TaskClient) checkAndCreateKernels() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Staring Check And create Kernels")

	kernelsInRedis, err := t.redisClient.Client.LLen(ctx, t.cfg.RedisKey).Result()
	if err != nil {
		log.Printf("[TASK:checkAndCreateKernels] Error when getting Redis list length: %v", err)
		return
	}

	creatingKernelCount := t.creatingKernelCount.Get()
	totalKernels := int(kernelsInRedis) + creatingKernelCount

	maxKernelLen := t.cfg.MaxPendingKernels

	if totalKernels < maxKernelLen {
		needCreateKernelCount := maxKernelLen - totalKernels
		log.Println("[TASK:checkAndCreateKernels] Check Result: needCreateKernelCount:", needCreateKernelCount)
		err := t.StartKernels(needCreateKernelCount)
		if err != nil {
			log.Printf("[TASK:checkAndCreateKernels] Error when starting kernels: %v", err)
		}
	} else {
		log.Println("[TASK:checkAndCreateKernels] No need to create more kernels")
	}

}

func (t *TaskClient) createKernel(reqBody map[string]interface{}) error {

	t.creatingKernelCount.Increment()
	defer t.creatingKernelCount.Decrement()

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		panic("cannot marshal reqBody,please check code.")
	}
	var kernelInfo models.KernelInfo
	var created bool
	created = false
	for i := 0; i < 3; i++ {
		err := func() error {
			// log.Printf("create kernel with json: %v", string(jsonData))
			resp, err := t.httpClient.Post("/api/kernels", jsonData)

			if err != nil {
				log.Printf("Failed to create kernel: %v,json data %v", err, string(jsonData))
				return err
			}

			dec := json.NewDecoder(bytes.NewReader(resp))
			dec.DisallowUnknownFields()

			err = dec.Decode(&kernelInfo)
			if err != nil {
				log.Printf("Failed to decode kernelInfo: %v,response: %v", err, string(resp))
				return err
			}
			return nil
		}()
		if err != nil {
			log.Printf("create kernel failed: %v,retry time: %v", err, i+1)
			time.Sleep(1 * time.Second)
			continue
		}
		created = true
		break
	}
	if !created {
		log.Println()
		return errors.New("cannot create kernel after 3 times")
	}

	log.Println("Created kernel:", kernelInfo)

	_, err = t.httpClient.Get("/api/kernels/" + kernelInfo.ID)

	if err != nil {
		log.Println("Cannot pre activate the kernel, id:", kernelInfo.ID)
	}
	log.Println("pre-activate after created done.")

	kernelJSON, err := json.Marshal(kernelInfo)
	if err != nil {
		// panic("Cannot Marshal kernelInfo!!!")
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = t.redisClient.Client.LPush(ctx, t.cfg.RedisKey, string(kernelJSON)).Err()
	if err != nil {
		// panic("Cannot LPush kernelInfo!!!")
		log.Println("Cannot LPush kernelInfo")
		return err
	}

	return nil

}

func (t *TaskClient) activateKernel(kernelId string) error {

	wsUrl := t.cfg.EGWSEndpoint + "/api/kernels/" + kernelId + "/channels"

	wsClient := common.NewWebSocketClient(wsUrl)
	defer wsClient.Close()

	for i := 0; i < 3; i++ {
		err := wsClient.Activate()

		if err != nil {
			log.Printf("Cannot connect to the websocket: %v,kernel ID: %v , retry count: %v", err, kernelId, i+1)
			log.Println("Do http get to EG to pre-activate the kernel,kernel ID", kernelId)

			_, err := t.httpClient.Get("/api/kernels/" + kernelId)

			if err != nil {
				log.Println("Cannot http get to activate the kernel,kernel id:", kernelId)
				time.Sleep(1 * time.Second)
			}

			log.Printf("pre-activate kernel ID: %v done", kernelId)
		} else {
			break
		}
	}

	idleCount := 0

	for {
		select {
		case message := <-wsClient.ResultChan:

			if common.InfoRequestResult(message, &idleCount) {
				log.Println("active the kernel done,ID: ", kernelId)
				return nil
			}
		case <-time.After(3 * time.Second):
			log.Printf("Waiting Timeout")
			return errors.New("waiting timeout")

		}

	}
}

// delete kernel by kernelID
func (t *TaskClient) deleteKernelByKernelId(kernelId string) error {
	log.Println("deleting kernel by kernel ID: ", kernelId)
	err := t.httpClient.Delete("/api/kernels/" + kernelId)
	if err != nil {
		log.Printf("Cannot delete kernel by kernel ID: %v", kernelId)
		return err
	}
	log.Println("Successfully delete kernel, kernel ID: ", kernelId)
	return nil
}

func (t *TaskClient) DeleteKernelByCount(needDeleteCount int) error {

	for i := 0; i < needDeleteCount; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var kernel models.KernelInfo
		kernelInfo, err := t.redisClient.Client.RPop(ctx, t.cfg.RedisKey).Result()
		if err != nil {
			continue
		}
		err = json.Unmarshal([]byte(kernelInfo), &kernel)
		if err != nil {
			log.Println("Cannot unmarshal the kernelInfo..")
		}

		err = t.deleteKernelByKernelId(kernel.ID)
		if err != nil {
			log.Printf("Cannot delete kernel by kernel ID: %v", kernel.ID)
			continue
		}
	}
	return nil

}
