package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"zjuici.com/tablegpt/jkpmanager/src/models"
	"zjuici.com/tablegpt/jkpmanager/src/storage"
)

type TaskClient struct {
	httpClient            *HTTPClient
	redisClient           *storage.RedisClient
	cfg                   *models.Config
	toCreateKernelsChan   chan map[string]interface{}
	toActivateKernelsChan chan string
	done                  chan struct{}
}

func NewTaskClient(cfg *models.Config) *TaskClient {
	return &TaskClient{
		toCreateKernelsChan:   make(chan map[string]interface{}, 200),
		toActivateKernelsChan: make(chan string, 200),
		cfg:                   cfg,
		done:                  make(chan struct{}),
		httpClient:            NewHTTPClient(cfg.EGEndpoint),
		redisClient:           storage.NewRedisClient(cfg.RedisHost, cfg.RedisPort),
	}
}

func (t *TaskClient) Start() {
	go t.startKernelsLoop()
	go t.activateKernelsLoop()
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

	kernelsJSON, err := t.redisClient.LRange(t.cfg.RedisKey, 0, -1)

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

func (t *TaskClient) createKernel(reqBody map[string]interface{}) error {

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		panic("cannot marshal reqBody,please check code.")
	}
	var kernelInfo models.KernelInfo
	var created bool
	created = false
	for i := 0; i < 3; i++ {
		err := func() error {
			log.Printf("create kernel with json: %v", string(jsonData))
			resp, err := t.httpClient.Post("/api/kernels", jsonData)

			if err != nil {
				log.Printf("Failed to create kernel: %v", err)
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

	// TODO: CREATED kernel then pre activate
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

	err = t.redisClient.LPush(t.cfg.RedisKey, string(kernelJSON))
	if err != nil {
		// panic("Cannot LPush kernelInfo!!!")
		log.Println("Cannot LPush kernelInfo")
		return err
	}

	return nil

}

func (t *TaskClient) activateKernel(kernelId string) error {

	wsUrl := t.cfg.EGWSEndpoint + "/api/kernels/" + kernelId + "/channels"

	wsClient := NewWebSocketClient(wsUrl)
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

			if InfoRequestResult(message, &idleCount) {
				log.Println("active the kernel done,ID: ", kernelId)
				return nil
			}
		case <-time.After(3 * time.Second):
			log.Printf("Waiting Timeout")
			return errors.New("waiting timeout")

		}

	}
}
