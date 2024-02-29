package common

import (
	"bytes"
	"encoding/json"
	"log"
	"sync"
	"time"

	"zjuici.com/tablegpt/jkpmanager/src/models"
	"zjuici.com/tablegpt/jkpmanager/src/storage"
)

func StartKernels(cfg *models.Config, httpClient *HTTPClient, redisClient *storage.RedisClient, needCreateKernelCount int) {

	// needCreateKernelCount := cfg.MaxPendingKernels - len(storedKernels)

	kernelVolumeMounts, err := json.Marshal([]map[string]string{
		{
			"name":      "shared-vol",
			"mountPath": cfg.WorkingDir,
		},
	})
	if err != nil {
		log.Println("Cannot marshal the kernelVolumeMounts")
	}
	kernelVolumes, err := json.Marshal([]map[string]interface{}{
		{"name": "shared-vol",
			"nfs": map[string]string{
				"server": cfg.NFSVolumeServer,
				"path":   cfg.NFSMountPath,
			},
		},
	})

	if err != nil {
		log.Println("Cannot marshal the kernelVolumes")
	}

	data := map[string]interface{}{
		"name": "python_kubernetes",
		"env": map[string]string{
			"KERNEL_NAMESPACE":     cfg.KernelNamespace,
			"KERNEL_WORKING_DIR":   cfg.WorkingDir,
			"KERNEL_VOLUME_MOUNTS": string(kernelVolumeMounts),
			"KERNEL_VOLUMES":       string(kernelVolumes),
			"KERNEL_IMAGE":         cfg.KernelImage,
		},
	}
	var wg sync.WaitGroup
	for i := 0; i < needCreateKernelCount; i++ {
		wg.Add(1)
		go createKernel(cfg, httpClient, redisClient, data, &wg)
	}

	wg.Wait()
}

func createKernel(cfg *models.Config, httpClient *HTTPClient, redisClient *storage.RedisClient, reqBody map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		panic("cannot marshal reqBody,please check code.")
	}
	var kernelInfo models.KernelInfo
	// 3 retry times
	var created bool
	created = false
	for i := 0; i < 3; i++ {
		err := func() error {
			resp, err := httpClient.Post("/api/kernels", jsonData)

			if err != nil {
				log.Printf("Failed to create kernel: %v", err)
				return err
			}

			dec := json.NewDecoder(bytes.NewReader(resp))
			dec.DisallowUnknownFields()

			err = dec.Decode(&kernelInfo)
			if err != nil {
				log.Printf("Failed to decode kernelInfo: %v", err)
				return err
			}
			return nil
		}()
		if err != nil {
			log.Printf("create kernel failed: %v,retry time: %v", err, i)
			time.Sleep(1 * time.Second)
			continue
		}
		created = true
		break
	}
	if !created {
		log.Println("cannot create kernel after 3 times.")
		return
	}
	kernelJSON, err := json.Marshal(kernelInfo)
	if err != nil {
		panic("Cannot Marshal kernelInfo!!!")
	}

	redisClient.LPush(cfg.RedisKey, string(kernelJSON))

}
