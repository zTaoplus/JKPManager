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
	EGEndpoint         string `mapstructure:"EG_ENDPOINT"`
	MaxPendingKernels  int    `mapstructure:"MAX_PENDING_KERNELS"`
	NFSVolumeServer    string `mapstructure:"NFS_VOLUME_SERVER"`
	NFSMountPath       string `mapstructure:"NFS_MOUNT_PATH"`
	WorkingDir         string `mapstructure:"WORKING_DIR"`
	KernelImage        string `mapstructure:"KERNEL_IMAGE"`
	KernelUserName     string `mapstructure:"KERNEL_USER_NAME"`
	ServerPort         string `mapstructure:"SERVER_PORT"`
	ActivationInterval int    `mapstructure:"ACTIVATION_INTERVAL"`
	CheckTaskInterval  int    `mapstructure:"CHECK_TASK_INTERVAL"`
	// CreateKernelThreshold float64 `mapstructure:"KERNEL_THRESHOLD"`
	RedisDSN string `mapstructure:"REDIS_DSN"`
	// RedisPort       string `mapstructure:"REDIS_PORT"`
	RedisDB           string `mapstructure:"REDIS_DB"`
	RedisKey          string `mapstructure:"REDIS_KEY"`
	KernelNamespace   string `mapstructure:"KERNEL_NAMESPACE"`
	EGWSEndpoint      string `mapstructure:"EG_WS_ENDPOINT"`
	KernelsSessionKey string `mapstructure:"KERNELS_SESSION_KEY"`

	DbType        string `mapstructure:"DB_TYPE"`
	PGDSN         string `mapstructure:"PG_DSN"`
	PGMaxPoolSize int    `mapstructure:"PG_MAX_POOL_SIZE"`
}
