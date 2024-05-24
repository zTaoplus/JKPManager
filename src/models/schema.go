package models

import "encoding/json"

// kernel info
type KernelInfo struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	LastActivity   string `json:"last_activity"`
	ExecutionState string `json:"execution_state"`
	Connections    int64  `json:"connections"`
}

type Session struct {
	ID string `json:"kernel_id"`
	// KernelInfo pgtype.JSONB `json:"kernel_session"`
	KernelInfo json.RawMessage `json:"kernel_session"`
}

type Config struct {
	// config about jupyter enterprise gateway
	EGEndpoint       string `mapstructure:"EG_ENDPOINT"`
	EGWebhookEnabled bool   `mapstructure:"EG_WEBHOOK_ENABLED"`
	EGWSEndpoint     string `mapstructure:"EG_WS_ENDPOINT"`

	// config about kernel create
	KernelName               string `mapstructure:"KERNEL_NAME"`
	KernelVolumeMounts       string `mapstructure:"KERNEL_VOLUME_MOUNTS"`
	KernelVolumes            string `mapstructure:"KERNEL_VOLUMES"`
	KernelStartupScriptsPath string `mapstructure:"KERNEL_STARTUP_SCRIPTS_PATH"`
	KernelWorkingDir         string `mapstructure:"KERNEL_WORKING_DIR"`
	KernelImage              string `mapstructure:"KERNEL_IMAGE"`
	KernelUserName           string `mapstructure:"KERNEL_USER_NAME"`
	KernelNamespace          string `mapstructure:"KERNEL_NAMESPACE"`

	// should add the config of the volume settings?

	// config about server
	ServerPort string `mapstructure:"SERVER_PORT"`

	// config about kernels operations
	MaxPendingKernels  int `mapstructure:"MAX_PENDING_KERNELS"`
	ActivationInterval int `mapstructure:"ACTIVATION_INTERVAL"`
	CheckTaskInterval  int `mapstructure:"CHECK_TASK_INTERVAL"`

	// db type to store sessions
	DbType string `mapstructure:"DB_TYPE"`

	// about redis storage
	RedisDSN          string `mapstructure:"REDIS_DSN"`
	RedisDB           string `mapstructure:"REDIS_DB"`
	RedisKey          string `mapstructure:"REDIS_KEY"`
	KernelsSessionKey string `mapstructure:"KERNELS_SESSION_KEY"`

	// about postgresql
	PGDSN         string `mapstructure:"PG_DSN"`
	PGMaxPoolSize int    `mapstructure:"PG_MAX_POOL_SIZE"`
}
