package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"zjuici.com/tablegpt/jkpmanager/src/common"
	"zjuici.com/tablegpt/jkpmanager/src/controllers"
	"zjuici.com/tablegpt/jkpmanager/src/scheduler"
	"zjuici.com/tablegpt/jkpmanager/src/storage"
)

func dbMiddleware(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := pool.Acquire(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer conn.Release()

			ctx := context.WithValue(r.Context(), common.DBConnKey, conn)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func main() {
	// init some config from env
	cfg, err := common.InitConfig()
	if err != nil {
		log.Panicf("cannot load config from env using viper, msg: %v", err)
	}

	// init router by mux
	r := mux.NewRouter()
	// init db
	err = storage.InitDBClient(cfg)
	dbClient := storage.GetDB()
	defer dbClient.Close()

	if err != nil {
		log.Panicln("Cannot init db client,err:", err)
	}

	// use db middleware
	r.Use(dbMiddleware(dbClient))

	// note(zt): hack http delete can buffer the request body
	r.Use(mux.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// 检查重定向
			if p := strings.TrimSuffix(req.URL.Path, "/"); p != req.URL.Path {
				// 如果发生了重定向，更新请求路径
				req.URL.Path = p
			}
			req.Body = http.MaxBytesReader(w, req.Body, 1048576) // 1 MB
			next.ServeHTTP(w, req)
		})
	}))

	// strict slash
	// r.StrictSlash(true)

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
	taskClient := scheduler.NewTaskClient(cfg)
	taskClient.Start()

	storedKernelsLen, err := redisClient.LLen(cfg.RedisKey)
	if err != nil {
		log.Panicf("Cannot get the kernel count from redis: %v", err)
	}
	log.Printf("Existing Pending Kernel Count: %v, Max Pending Kernel Count: %v", storedKernelsLen, cfg.MaxPendingKernels)

	needCreateKernelCount := cfg.MaxPendingKernels - int(storedKernelsLen)
	if needCreateKernelCount < 0 {
		log.Println("need to delete :", -needCreateKernelCount)
		taskClient.DeleteKernelByCount(-needCreateKernelCount)

		needCreateKernelCount = 0
	}

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

	// start http server then create kernels
	go func() {
		log.Println("Staring http server")
		http.Handle("/", r)
		// eg kernels management
		r.HandleFunc("/api/kernels/pop/", controllers.PopKernelHandler(cfg, taskClient, redisClient)).Methods("POST")

		// eg sessions
		r.HandleFunc("/api/kernels/sessions/", controllers.GetKernelsHandler).Methods(http.MethodGet)                // get kernels
		r.HandleFunc("/api/kernels/sessions", controllers.GetKernelsHandler).Methods(http.MethodGet)                 // get kernels
		r.HandleFunc("/api/kernels/sessions/{kernelId}", controllers.GetKernelByIdHandler).Methods(http.MethodGet)   // get kernel about kernel id
		r.HandleFunc("/api/kernels/sessions/{kernelId}", controllers.PostKernelByIdHandler).Methods(http.MethodPost) // save kernels
		r.HandleFunc("/api/kernels/sessions/", controllers.DeleteKernelsHandler).Methods(http.MethodDelete)          // delete kernels by ids
		r.HandleFunc("/api/kernels/sessions", controllers.DeleteKernelsHandler).Methods(http.MethodDelete)           // delete kernels by ids

		if err := http.ListenAndServe(":"+cfg.ServerPort, nil); err != nil {
			log.Panicln("Error starting HTTP server:", err)
		}
	}()

	// starting kernel create task
	log.Println("needCreateKernelCount:", needCreateKernelCount)
	taskClient.StartKernels(needCreateKernelCount)

	//
	select {}
}
