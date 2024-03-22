package common

import (
	"log"
	"strings"

	"github.com/spf13/viper"
	"zjuici.com/tablegpt/jkpmanager/src/models"
)

type contextKey string

const SessionClientKey contextKey = "session-client"

var Cfg *models.Config

func InitConfig() error {
	viper.SetEnvPrefix("JKP")
	viper.AutomaticEnv()

	viper.SetDefault("EG_ENDPOINT", "http://127.0.0.1:8888")
	viper.SetDefault("MAX_PENDING_KERNELS", 10)
	viper.SetDefault("NFS_VOLUME_SERVER", "127.0.0.1")
	viper.SetDefault("NFS_MOUNT_PATH", "/data/")
	viper.SetDefault("WORKING_DIR", "/mnt/shared")
	viper.SetDefault("KERNEL_IMAGE", "elyra/kernel-py:3.2.2")
	viper.SetDefault("KERNEL_NAMESPACE", "default")
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("ACTIVATION_INTERVAL", 1800)
	viper.SetDefault("CHECK_TASK_INTERVAL", 120)
	viper.SetDefault("REDIS_DSN", "redis://127.0.0.1:6379")
	viper.SetDefault("REDIS_KEY", "jupyter:kernels:idle")

	viper.SetDefault("KERNELS_SESSION_KEY", "jupyter:kernels:sessions")

	viper.SetDefault("DB_TYPE", "redis")

	viper.SetDefault("PG_DSN", "postgresql://postgres:zjuici@127.0.0.1:5432/postgres?search_path=public")
	viper.SetDefault("PG_MAX_POOL_SIZE", "20")

	err := viper.Unmarshal(&Cfg)
	if err != nil {
		return err
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
