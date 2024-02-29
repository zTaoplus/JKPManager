package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"zjuici.com/tablegpt/jpkmanager/src/common"
	"zjuici.com/tablegpt/jpkmanager/src/controllers"
	"zjuici.com/tablegpt/jpkmanager/src/storage"
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

	httpClient := common.NewHTTPClient(cfg.EGEndpoint)

	redisClient := storage.NewRedisClient(cfg.RedisHost, cfg.RedisPort)
	// init kernel

	// if needCreateKernelCount
	storedKernels, err := redisClient.LRange(cfg.RedisKey, 0, -1)
	if err != nil {
		panic(err)
	}

	needCreateKernelCount := cfg.MaxPendingKernels - len(storedKernels)
	log.Printf("Existing Pending Kernel Count: %v, needCreateKernelCount: %v", len(storedKernels), needCreateKernelCount)
	common.StartKernels(cfg, httpClient, redisClient, needCreateKernelCount)

	// 启动web监听
	go func() {
		log.Println("Staring http server")
		http.Handle("/", r)
		r.HandleFunc("/api/kernels/pop/", controllers.PopKernelHandler(cfg, httpClient, redisClient)).Methods("POST")
		log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, nil))
	}()

	// 启动定时任务
	go func() {
		log.Printf("Start the scheduled task KernelActivator, activate at intervals of %v seconds.", cfg.ActivationInterval)

		ticker := time.NewTicker(time.Duration(cfg.ActivationInterval) * time.Second)

		// task
		for range ticker.C {
			log.Println("原神！ 启动！")
			common.KernelActivateTask(cfg, redisClient)
		}
	}()

	// 阻塞主goroutine
	select {}

}