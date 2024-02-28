package models

import "time"

// kernel info
type KernelInfo struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	LastActivity   string `json:"last_activity"`
	ExecutionState string `json:"execution_state"`
	Connections    int64  `json:"connections"`
}

// pop request body
type OwnerUser struct {
	ID string `json:"userId"`
}

type EgWSReplyHeader struct {
	MsgID    string    `json:"msg_id"`
	MsgType  string    `json:"msg_type"`
	Username string    `json:"username"`
	Session  string    `json:"session"`
	Date     time.Time `json:"date"`
	Version  string    `json:"version"`
}

type EgWSReplyParentHeader struct {
	MsgID    string    `json:"msg_id"`
	MsgType  string    `json:"msg_type"`
	Username string    `json:"username"`
	Session  string    `json:"session"`
	Date     time.Time `json:"date"`
	Version  string    `json:"version"`
}

type EgWsReplyStatus struct {
	ExecutionState string `json:"execution_state"`
}

type EgWsReplyExecuteInput struct {
	Code           string `json:"code"`
	ExecutionCount int    `json:"execution_count"`
}

type EgWsReplyExecuteReply struct {
	Status          string                 `json:"status"`
	ExecutionCount  int                    `json:"execution_count"`
	UserExpressions map[string]interface{} `json:"user_expressions"`
	Payload         []interface{}          `json:"payload"`
}

type EgWsReplyExecuteResult struct {
	Data           interface{} `json:"data"`
	Metadata       interface{} `json:"metadata"`
	ExecutionCount int         `json:"execution_count"`
}

type EgWSReplyMessage struct {
	Header       EgWSReplyHeader        `json:"header"`
	MsgID        string                 `json:"msg_id"`
	MsgType      string                 `json:"msg_type"`
	ParentHeader EgWSReplyParentHeader  `json:"parent_header"`
	Metadata     interface{}            `json:"metadata"`
	Content      map[string]interface{} `json:"content"`

	Buffers []interface{} `json:"buffers"`
	Channel string        `json:"channel"`
}

type EgWsSendContent struct {
	Code            string                 `json:"code"`
	Silent          bool                   `json:"silent"`
	StoreHistory    bool                   `json:"store_history"`
	UserExpressions map[string]interface{} `json:"user_expressions"`
	AllowStdin      bool                   `json:"allow_stdin"`
}

type EgWsSendHeader struct {
	MsgID   string `json:"msg_id"`
	MsgType string `json:"msg_type"`
}

type EgWsSendMessage struct {
	Header       EgWsSendHeader    `json:"header"`
	ParentHeader map[string]string `json:"parent_header"`
	Metadata     map[string]string `json:"metadata"`
	Content      EgWsSendContent   `json:"content"`
	Channel      string            `json:"channel"`
}

type Config struct {
	EGEndpoint         string `mapstructure:"EG_ENDPOINT"`
	MaxPendingKernels  int    `mapstructure:"MAX_PENDING_KERNELS"`
	NFSVolumeServer    string `mapstructure:"NFS_VOLUME_SERVER"`
	NFSMountPath       string `mapstructure:"NFS_MOUNT_PATH"`
	WorkingDir         string `mapstructure:"WORKING_DIR"`
	KernelImage        string `mapstructure:"KERNEL_IMAGE"`
	ServerPort         string `mapstructure:"SERVER_PORT"`
	ActivationInterval int    `mapstructure:"ACTIVATION_INTERVAL"`
	// CreateKernelThreshold float64 `mapstructure:"KERNEL_THRESHOLD"`
	RedisHost       string `mapstructure:"REDIS_HOST"`
	RedisPort       string `mapstructure:"REDIS_PORT"`
	RedisDB         string `mapstructure:"REDIS_DB"`
	RedisKey        string `mapstructure:"REDIS_KEY"`
	KernelNamespace string `mapstructure:"KERNEL_NAMESPACE"`
	EGWSEndpoint    string `mapstructure:"EG_WS_ENDPOINT"`
}
