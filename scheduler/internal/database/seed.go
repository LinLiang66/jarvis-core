package database

import (
	"context"
	"log"

	"jarvis-core/scheduler/internal/model"
	"jarvis-core/scheduler/internal/store"
)

const (
	demoJobName    = "示例-Hello定时任务"
	demoJobHandler = "demo.hello"
)

// Seed 幂等写入示例定时任务（同名或同 handler 已存在则跳过）。
func Seed(ctx context.Context, s *store.Stores) error {
	exists, err := s.JobExistsByNameOrHandler(ctx, demoJobName, demoJobHandler)
	if err != nil {
		return err
	}
	if exists {
		log.Printf("[seed] demo job already exists, skip name=%s handler=%s", demoJobName, demoJobHandler)
		return nil
	}

	job := model.JobDefinition{
		GroupName:     "default",
		Name:          demoJobName,
		Handler:       demoJobHandler,
		TriggerType:   model.TriggerCron,
		CronExpr:      "0 */5 * * * *",
		Params:        `{}`,
		BlockStrategy: model.BlockSerial,
		RouteStrategy: model.RouteRoundRobin,
		ExecuteMode:   model.ExecuteCluster,
		Status:        model.StatusEnabled,
		Description:   "示例任务，每5分钟打印日志",
		TimeoutSec:    60,
		RetryCount:    0,
		RetryInterval: 60,
		ParallelCount: 1,
	}
	if err := s.CreateJob(ctx, &job); err != nil {
		return err
	}
	log.Printf("[seed] created demo job id=%d name=%s handler=%s cron=%s", job.ID, job.Name, job.Handler, job.CronExpr)
	return nil
}
