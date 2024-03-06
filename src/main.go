package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"zjuici.com/tablegpt/jkpmanager/src/common"
	"zjuici.com/tablegpt/jkpmanager/src/controllers"
	"zjuici.com/tablegpt/jkpmanager/src/storage"
)

func main() {
	// init some config from env
	cfg, err := common.InitConfig()
	if err != nil {
		log.Panicf("cannot load config from env using viper, msg: %v", err)
	}

	// init router by mux
	r := mux.NewRouter()

	// strict slash
	r.StrictSlash(true)

	// httpClient := common.NewHTTPClient(cfg.EGEndpoint)

	redisClient := storage.NewRedisClient(cfg.RedisHost, cfg.RedisPort)

	// check redis health
	redisHealth := false
	for i := 0; i < 5; i++ {
		err := redisClient.Ping()
		if err != nil {
			log.Printf("Failed to ping redis, retry count: %v", i+1)
			time.Sleep(1500 * time.Millisecond)
			continue
		}
		redisHealth = true
	}

	if !redisHealth {
		log.Panicf("Failed to ping redis after 5 retries,error: %v", err)
	}

	// start task
	taskClient := common.NewTaskClient(cfg)
	taskClient.Start()

	storedKernelsLen, err := redisClient.LLen(cfg.RedisKey)
	if err != nil {
		log.Panicf("Cannot get the kernel count from redis: %v", err)
	}

	// TODO: when needCreateKernelCount is <0, we will pop kernels in redis and delete it by eg url delete api.
	needCreateKernelCount := cfg.MaxPendingKernels - int(storedKernelsLen)
	if needCreateKernelCount < 0 {
		log.Println("need to delete :", -needCreateKernelCount)
		taskClient.DeleteKernelByCount(-needCreateKernelCount)

		needCreateKernelCount = 0
	}
	log.Printf("Existing Pending Kernel Count: %v, needCreateKernelCount: %v", storedKernelsLen, needCreateKernelCount)
	taskClient.StartKernels(needCreateKernelCount)

	// 启动定时任务
	go func() {
		log.Printf("Start the scheduled task KernelActivator, activate at intervals of %v seconds.", cfg.ActivationInterval)

		ticker := time.NewTicker(time.Duration(cfg.ActivationInterval) * time.Second)

		// task
		for range ticker.C {
			log.Println("Scheduled task starting!")
			taskClient.ActivateKernels()
		}
	}()

	log.Println("Staring http server")
	http.Handle("/", r)
	r.HandleFunc("/api/kernels/pop/", controllers.PopKernelHandler(cfg, taskClient, redisClient)).Methods("POST")
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, nil))

}
