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
	// init kernel

	taskClient := common.NewTaskClient(cfg)
	taskClient.Start()

	// if needCreateKernelCount
	storedKernelsLen, err := redisClient.LLen(cfg.RedisKey)
	if err != nil {
		panic(err)
	}

	needCreateKernelCount := cfg.MaxPendingKernels - int(storedKernelsLen)
	if needCreateKernelCount < 0 {
		needCreateKernelCount = 0
	}
	log.Printf("Existing Pending Kernel Count: %v, needCreateKernelCount: %v", storedKernelsLen, needCreateKernelCount)
	taskClient.StartKernels(needCreateKernelCount)

	// 启动web监听
	go func() {
		log.Println("Staring http server")
		http.Handle("/", r)
		r.HandleFunc("/api/kernels/pop/", controllers.PopKernelHandler(cfg, taskClient, redisClient)).Methods("POST")
		log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, nil))
	}()

	// 启动定时任务
	go func() {
		log.Printf("Start the scheduled task KernelActivator, activate at intervals of %v seconds.", cfg.ActivationInterval)

		ticker := time.NewTicker(time.Duration(cfg.ActivationInterval) * time.Second)

		// task
		for range ticker.C {
			log.Println("Scheduled task starting!")
			go taskClient.ActivateKernels()
		}
	}()

	// 阻塞主goroutine
	select {}

}
