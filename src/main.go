package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"zjuici.com/tablegpt/jkpmanager/src/common"
	"zjuici.com/tablegpt/jkpmanager/src/controllers"
	"zjuici.com/tablegpt/jkpmanager/src/scheduler"
	"zjuici.com/tablegpt/jkpmanager/src/storage"
)

func sessionClientMiddleware(client storage.SessionClient) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), common.SessionClientKey, client)
			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}

func egWebHookRouter(r *mux.Router) {
	var client storage.SessionClient

	var err error

	switch common.Cfg.DbType {
	case "postgres":
		err := storage.InitDBClient()
		if err != nil {
			log.Panicln("Cannot init db client for postgres,err:", err)
		}
		client = storage.GetDB()
	default:

		if err != nil {
			log.Panicln("Cannot init redis client,err:", err)
		}
		client = storage.GetRedisClient()
	}

	r.Use(sessionClientMiddleware(client))

	// note(zt): hack http delete can buffer the request body
	r.Use(mux.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			req.Body = http.MaxBytesReader(w, req.Body, 1048576) // 1 MB
			next.ServeHTTP(w, req)
		})
	}))

	r.HandleFunc("/api/kernels/sessions/", controllers.GetKernelsHandler).Methods(http.MethodGet)                // get kernels
	r.HandleFunc("/api/kernels/sessions", controllers.GetKernelsHandler).Methods(http.MethodGet)                 // get kernels
	r.HandleFunc("/api/kernels/sessions/{kernelId}", controllers.GetKernelByIdHandler).Methods(http.MethodGet)   // get kernel about kernel id
	r.HandleFunc("/api/kernels/sessions/{kernelId}", controllers.PostKernelByIdHandler).Methods(http.MethodPost) // save kernels
	r.HandleFunc("/api/kernels/sessions/", controllers.DeleteKernelsHandler).Methods(http.MethodDelete)          // delete kernels by ids
	r.HandleFunc("/api/kernels/sessions", controllers.DeleteKernelsHandler).Methods(http.MethodDelete)           // delete kernels by ids
}

func main() {
	// init some config from env
	err := common.InitConfig()
	if err != nil {
		log.Panicf("cannot load config from env using viper, msg: %v", err)
	}
	// init redis client

	err = storage.InitRedisClient()
	if err != nil {
		log.Panicln("Cannot init redis Client")
	}

	// init router by mux
	r := mux.NewRouter()

	// eg sessions enabled
	if common.Cfg.EGWebhookEnabled {
		log.Println("EG Webhook Enabled")
		egWebHookRouter(r)
	}

	// Opening the strict slash will prevent the delete method from saving the request body.
	// r.StrictSlash(true)

	// start task
	taskClient := scheduler.NewTaskClient(common.Cfg)
	taskClient.Start()

	// start http server then create kernels
	go func() {
		log.Println("Staring http server")
		http.Handle("/", r)
		// eg kernels management
		r.HandleFunc("/api/kernels/pop/", controllers.PopKernelHandler(common.Cfg, taskClient)).Methods(http.MethodPost)
		loggedRouter := handlers.LoggingHandler(os.Stdout, r)

		if err := http.ListenAndServe(":"+common.Cfg.ServerPort, loggedRouter); err != nil {
			log.Panicln("Error starting HTTP server:", err)
		}

	}()

	taskClient.InitKernels()
	go taskClient.ActivateKernels()

	select {}
}
