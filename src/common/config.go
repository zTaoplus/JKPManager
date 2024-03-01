package common

import (
	"log"
	"strings"

	"github.com/spf13/viper"
	"zjuici.com/tablegpt/jkpmanager/src/models"
)

func InitConfig() (*models.Config, error) {
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
	viper.SetDefault("REDIS_HOST", "127.0.0.1")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_DB", "0")
	viper.SetDefault("REDIS_KEY", "jupyter:kernels:idle")

	var cfg models.Config

	err := viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}
	var egWsEndpoint string

	if strings.HasPrefix(cfg.EGEndpoint, "http://") {
		egWsEndpoint = strings.Replace(cfg.EGEndpoint, "http://", "ws://", 1)
	} else if strings.HasPrefix(cfg.EGEndpoint, "https://") {
		egWsEndpoint = strings.Replace(cfg.EGEndpoint, "https://", "wss://", 1)
	} else {
		log.Printf("invalid protocol endpoint：%v", cfg.EGEndpoint)
		panic("cannot parse egEndpoint to egWsEndpoint")
	}

	cfg.EGWSEndpoint = egWsEndpoint
	return &cfg, nil
}
