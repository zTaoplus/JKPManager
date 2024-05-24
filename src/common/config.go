package common

import (
	"log"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"zjuici.com/tablegpt/jkpmanager/src/models"
)

type contextKey string

const SessionClientKey contextKey = "session-client"

var Cfg = &models.Config{}

func InitConfig() error {
	viper.SetEnvPrefix("JKP")
	viper.AutomaticEnv()
	// FIXME:Some config settings have default values that are not necessary, but without them, viper cannot read them.
	// server config
	viper.SetDefault("SERVER_PORT", "8080")

	// EG config
	viper.SetDefault("EG_ENDPOINT", "http://127.0.0.1:8888")
	viper.SetDefault("EG_WEBHOOK_ENABLED", false)

	// kernel config
	viper.SetDefault("KERNEL_VOLUME_MOUNTS", "")
	viper.SetDefault("KERNEL_VOLUMES", "")
	viper.SetDefault("KERNEL_STARTUP_SCRIPTS_PATH", "")
	viper.SetDefault("KERNEL_NAME", "python_kubernetes")
	viper.SetDefault("KERNEL_IMAGE", "elyra/kernel-py:3.2.2")
	viper.SetDefault("KERNEL_NAMESPACE", "default")
	viper.SetDefault("KERNEL_USER_NAME", "jovyan")

	// task config
	viper.SetDefault("MAX_PENDING_KERNELS", 10)
	viper.SetDefault("ACTIVATION_INTERVAL", 1800)
	viper.SetDefault("CHECK_TASK_INTERVAL", 120)

	// redis config
	viper.SetDefault("REDIS_DSN", "redis://127.0.0.1:6379")
	viper.SetDefault("REDIS_KEY", "jupyter:kernels:idle")

	// webhook config
	viper.SetDefault("KERNELS_SESSION_KEY", "jupyter:kernels:sessions")
	viper.SetDefault("DB_TYPE", "redis")
	viper.SetDefault("PG_DSN", "postgresql://postgres:zjuici@127.0.0.1:5432/postgres?search_path=public")
	viper.SetDefault("PG_MAX_POOL_SIZE", "20")

	err := viper.Unmarshal(Cfg)
	if err != nil {
		return err
	}

	// Parse EG_WEBHOOK_ENABLED environment variable
	egWebhookStr := viper.GetString("EG_WEBHOOK_ENABLED")
	egWebhookEnabled, err := strconv.ParseBool(egWebhookStr)
	if err != nil {
		Cfg.EGWebhookEnabled = false
	} else {
		Cfg.EGWebhookEnabled = egWebhookEnabled
	}

	var egWsEndpoint string

	if strings.HasPrefix(Cfg.EGEndpoint, "http://") {
		egWsEndpoint = strings.Replace(Cfg.EGEndpoint, "http://", "ws://", 1)
	} else if strings.HasPrefix(Cfg.EGEndpoint, "https://") {
		egWsEndpoint = strings.Replace(Cfg.EGEndpoint, "https://", "wss://", 1)
	} else {
		log.Printf("invalid protocol endpoint：%v", Cfg.EGEndpoint)
		panic("cannot parse egEndpoint to egWsEndpoint")
	}

	Cfg.EGWSEndpoint = egWsEndpoint
	return nil
}
