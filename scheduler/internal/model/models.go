package model

import "time"

const (
	BlockParallel = "parallel"
	BlockSerial   = "serial"
	BlockDiscard  = "discard"

	RouteRoundRobin = "round_robin"
	ExecuteCluster  = "cluster"

	StatusEnabled  = "0"
	StatusDisabled = "1"

	TriggerCron       = "cron"
	TriggerFixedRate  = "fixed_rate"
	TriggerManual     = "manual"

	InstPending   = "PENDING"
	InstQueued    = "QUEUED"
	InstRunning   = "RUNNING"
	InstSuccess   = "SUCCESS"
	InstFailed    = "FAILED"
	InstDiscarded = "DISCARDED"

	WorkerOnline  = "online"
	WorkerOffline = "offline"
)

type JobDefinition struct {
	ID             int64     `json:"id" gorm:"primaryKey"`
	GroupName      string    `json:"group_name" gorm:"size:64;default:default;index"`
	Name           string    `json:"name" gorm:"size:128;not null"`
	Handler        string    `json:"handler" gorm:"size:64;not null;index"`
	TriggerType    string    `json:"trigger_type" gorm:"size:16;default:cron"`
	CronExpr       string    `json:"cron_expr" gorm:"size:64"`
	Params         string    `json:"params" gorm:"type:text"`
	BlockStrategy  string    `json:"block_strategy" gorm:"size:32;default:parallel"`
	RouteStrategy  string    `json:"route_strategy" gorm:"size:32;default:round_robin"`
	ExecuteMode    string    `json:"execute_mode" gorm:"size:32;default:cluster"`
	Status         string    `json:"status" gorm:"size:8;default:0;index"`
	Description    string    `json:"description" gorm:"size:512"`
	TimeoutSec     int       `json:"timeout_sec" gorm:"default:300"`
	RetryCount     int       `json:"retry_count" gorm:"default:0"`
	RetryInterval  int       `json:"retry_interval" gorm:"default:60"`
	ParallelCount  int       `json:"parallel_count" gorm:"default:1"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (JobDefinition) TableName() string { return "job_definition" }

type JobInstance struct {
	ID          int64      `json:"id" gorm:"primaryKey;index:idx_inst_poll,priority:3"`
	JobID       int64      `json:"job_id" gorm:"index;not null"`
	JobName     string     `json:"job_name" gorm:"size:128"`
	Handler     string     `json:"handler" gorm:"size:64;index:idx_inst_poll,priority:2"`
	TriggerType string     `json:"trigger_type" gorm:"size:16"`
	Status      string     `json:"status" gorm:"size:16;index:idx_inst_poll,priority:1"`
	WorkerID    string     `json:"worker_id" gorm:"size:64;index"`
	Params      string     `json:"params" gorm:"type:text"`
	TimeoutSec  int        `json:"timeout_sec" gorm:"default:300"`
	Result      string     `json:"result" gorm:"type:text"`
	ErrorMsg    string     `json:"error_msg" gorm:"type:text"`
	StartedAt   *time.Time `json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (JobInstance) TableName() string { return "job_instance" }

type JobLog struct {
	ID         int64     `json:"id" gorm:"primaryKey"`
	InstanceID int64     `json:"instance_id" gorm:"index;not null"`
	Level      string    `json:"level" gorm:"size:16"`
	Message    string    `json:"message" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at"`
}

func (JobLog) TableName() string { return "job_log" }

type WorkerNode struct {
	ID              int64     `json:"id" gorm:"primaryKey"`
	WorkerID        string    `json:"worker_id" gorm:"size:64;uniqueIndex;not null"`
	InstanceID      string    `json:"instance_id" gorm:"size:64;index"`
	Hostname        string    `json:"hostname" gorm:"size:128"`
	Handlers        string    `json:"handlers" gorm:"type:text"`
	Status          string    `json:"status" gorm:"size:16;index"`
	LastHeartbeatAt *time.Time `json:"last_heartbeat_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (WorkerNode) TableName() string { return "worker_node" }
