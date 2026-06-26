package store

import (
	"context"
	"encoding/json"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	mysqlcfg "jarvis-core/scheduler/internal/infra/mysql"
	"jarvis-core/scheduler/internal/model"
)

type Stores struct {
	DB *gorm.DB
}

func Open(dsn string, mysqlCfg mysqlcfg.Config) (*Stores, error) {
	logMode := logger.Warn
	if mysqlCfg.ShowSQL {
		logMode = logger.Info
	}
	if mysqlCfg.LogLevel > 0 {
		logMode = logger.LogLevel(mysqlCfg.LogLevel)
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logMode),
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if mysqlCfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(mysqlCfg.MaxOpenConns)
	}
	if mysqlCfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(mysqlCfg.MaxIdleConns)
	}
	s := &Stores{DB: db}
	if err := s.AutoMigrate(context.Background()); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Stores) AutoMigrate(ctx context.Context) error {
	return s.DB.WithContext(ctx).AutoMigrate(
		&model.JobDefinition{},
		&model.JobInstance{},
		&model.JobLog{},
		&model.WorkerNode{},
	)
}

type PageQuery struct {
	Page  int
	Size  int
	Order string
}

func normalizePage(page, size int) (offset, limit int) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	return (page - 1) * size, size
}

type JobFilter struct {
	Name    string
	Handler string
	Status  string
}

func (s *Stores) ListJobs(ctx context.Context, pq PageQuery, f JobFilter) ([]model.JobDefinition, int64, error) {
	q := s.DB.WithContext(ctx).Model(&model.JobDefinition{})
	if f.Name != "" {
		q = q.Where("name LIKE ?", "%"+f.Name+"%")
	}
	if f.Handler != "" {
		q = q.Where("handler = ?", f.Handler)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	order := pq.Order
	if order == "" {
		order = "id desc"
	}
	offset, limit := normalizePage(pq.Page, pq.Size)
	var list []model.JobDefinition
	if err := q.Order(order).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *Stores) GetJob(ctx context.Context, id int64) (*model.JobDefinition, error) {
	var job model.JobDefinition
	if err := s.DB.WithContext(ctx).First(&job, id).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *Stores) CreateJob(ctx context.Context, job *model.JobDefinition) error {
	return s.DB.WithContext(ctx).Create(job).Error
}

func (s *Stores) JobExistsByNameOrHandler(ctx context.Context, name, handler string) (bool, error) {
	var count int64
	err := s.DB.WithContext(ctx).Model(&model.JobDefinition{}).
		Where("name = ? OR handler = ?", name, handler).
		Count(&count).Error
	return count > 0, err
}

func (s *Stores) UpdateJob(ctx context.Context, job *model.JobDefinition) error {
	return s.DB.WithContext(ctx).Save(job).Error
}

func (s *Stores) DeleteJobs(ctx context.Context, ids []int64) error {
	return s.DB.WithContext(ctx).Delete(&model.JobDefinition{}, ids).Error
}

func (s *Stores) ListEnabledJobs(ctx context.Context) ([]model.JobDefinition, error) {
	var list []model.JobDefinition
	err := s.DB.WithContext(ctx).Where("status = ?", model.StatusEnabled).Find(&list).Error
	return list, err
}

type InstanceFilter struct {
	JobID   int64
	Handler string
	Status  string
}

func (s *Stores) ListInstances(ctx context.Context, pq PageQuery, f InstanceFilter) ([]model.JobInstance, int64, error) {
	q := s.DB.WithContext(ctx).Model(&model.JobInstance{})
	if f.JobID > 0 {
		q = q.Where("job_id = ?", f.JobID)
	}
	if f.Handler != "" {
		q = q.Where("handler = ?", f.Handler)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	order := pq.Order
	if order == "" {
		order = "id desc"
	}
	offset, limit := normalizePage(pq.Page, pq.Size)
	var list []model.JobInstance
	if err := q.Order(order).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *Stores) GetInstance(ctx context.Context, id int64) (*model.JobInstance, error) {
	var inst model.JobInstance
	if err := s.DB.WithContext(ctx).First(&inst, id).Error; err != nil {
		return nil, err
	}
	return &inst, nil
}

func (s *Stores) CreateInstance(ctx context.Context, inst *model.JobInstance) error {
	return s.DB.WithContext(ctx).Create(inst).Error
}

func (s *Stores) UpdateInstance(ctx context.Context, inst *model.JobInstance) error {
	return s.DB.WithContext(ctx).Save(inst).Error
}

// ClaimPendingInstance 查找最早待执行任务并以原子 UPDATE 认领，RowsAffected=0 表示竞争失败。
func (s *Stores) ClaimPendingInstance(ctx context.Context, workerID string, handlers []string) (*model.JobInstance, error) {
	if len(handlers) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var candidate model.JobInstance
	err := s.DB.WithContext(ctx).
		Where("status = ? AND handler IN ?", model.InstPending, handlers).
		Order("id asc").
		First(&candidate).Error
	if err != nil {
		return nil, err
	}
	result := s.DB.WithContext(ctx).Model(&model.JobInstance{}).
		Where("id = ? AND status = ? AND (worker_id = '' OR worker_id = ?)", candidate.ID, model.InstPending, workerID).
		Update("worker_id", workerID)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	candidate.WorkerID = workerID
	return &candidate, nil
}

func (s *Stores) AppendLog(ctx context.Context, instanceID int64, level, message string) error {
	return s.DB.WithContext(ctx).Create(&model.JobLog{
		InstanceID: instanceID,
		Level:      level,
		Message:    message,
	}).Error
}

func (s *Stores) ListLogs(ctx context.Context, instanceID int64, pq PageQuery) ([]model.JobLog, int64, error) {
	q := s.DB.WithContext(ctx).Model(&model.JobLog{}).Where("instance_id = ?", instanceID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset, limit := normalizePage(pq.Page, pq.Size)
	var list []model.JobLog
	if err := q.Order("id asc").Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *Stores) UpsertWorker(ctx context.Context, w *model.WorkerNode) error {
	var existing model.WorkerNode
	err := s.DB.WithContext(ctx).Where("worker_id = ?", w.WorkerID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return s.DB.WithContext(ctx).Create(w).Error
	}
	if err != nil {
		return err
	}
	w.ID = existing.ID
	return s.DB.WithContext(ctx).Save(w).Error
}

func (s *Stores) TouchWorker(ctx context.Context, workerID string, at time.Time) error {
	return s.DB.WithContext(ctx).Model(&model.WorkerNode{}).
		Where("worker_id = ?", workerID).
		Updates(map[string]any{
			"last_heartbeat_at": at,
			"status":            model.WorkerOnline,
		}).Error
}

func (s *Stores) ListWorkers(ctx context.Context, pq PageQuery) ([]model.WorkerNode, int64, error) {
	q := s.DB.WithContext(ctx).Model(&model.WorkerNode{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset, limit := normalizePage(pq.Page, pq.Size)
	var list []model.WorkerNode
	if err := q.Order("id desc").Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *Stores) ListEnabledJobHandlers(ctx context.Context) (map[string]struct{}, error) {
	var handlers []string
	err := s.DB.WithContext(ctx).Model(&model.JobDefinition{}).
		Where("status = ?", model.StatusEnabled).
		Distinct("handler").
		Pluck("handler", &handlers).Error
	if err != nil {
		return nil, err
	}
	set := make(map[string]struct{}, len(handlers))
	for _, h := range handlers {
		if h != "" {
			set[h] = struct{}{}
		}
	}
	return set, nil
}

func (s *Stores) ListPendingHandlers(ctx context.Context) (map[string]struct{}, error) {
	var handlers []string
	err := s.DB.WithContext(ctx).Model(&model.JobInstance{}).
		Where("status = ?", model.InstPending).
		Distinct("handler").
		Pluck("handler", &handlers).Error
	if err != nil {
		return nil, err
	}
	set := make(map[string]struct{}, len(handlers))
	for _, h := range handlers {
		if h != "" {
			set[h] = struct{}{}
		}
	}
	return set, nil
}

func (s *Stores) CollectOnlineWorkerHandlers(ctx context.Context) (map[string]struct{}, error) {
	var workers []model.WorkerNode
	if err := s.DB.WithContext(ctx).Where("status = ?", model.WorkerOnline).Find(&workers).Error; err != nil {
		return nil, err
	}
	set := make(map[string]struct{})
	for _, w := range workers {
		for _, h := range parseWorkerHandlers(w.Handlers) {
			if h != "" {
				set[h] = struct{}{}
			}
		}
	}
	return set, nil
}

func parseWorkerHandlers(handlersJSON string) []string {
	if handlersJSON == "" {
		return nil
	}
	var list []string
	if err := json.Unmarshal([]byte(handlersJSON), &list); err != nil {
		return nil
	}
	return list
}
func (s *Stores) ListStaleClaimedPending(ctx context.Context, cutoff time.Time) ([]model.JobInstance, error) {
	var list []model.JobInstance
	err := s.DB.WithContext(ctx).
		Where("status = ? AND worker_id != '' AND updated_at < ?", model.InstPending, cutoff).
		Find(&list).Error
	return list, err
}

func (s *Stores) ReclaimInstance(ctx context.Context, id int64, workerID string) (bool, error) {
	result := s.DB.WithContext(ctx).Model(&model.JobInstance{}).
		Where("id = ? AND status = ? AND worker_id = ?", id, model.InstPending, workerID).
		Update("worker_id", "")
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (s *Stores) ListTimedOutRunning(ctx context.Context, now time.Time) ([]model.JobInstance, error) {
	var list []model.JobInstance
	err := s.DB.WithContext(ctx).
		Where("status = ? AND started_at IS NOT NULL", model.InstRunning).
		Where("DATE_ADD(started_at, INTERVAL timeout_sec SECOND) < ?", now).
		Find(&list).Error
	return list, err
}

func (s *Stores) ListUndispatchedByHandlers(ctx context.Context, handlers []string) ([]model.JobInstance, error) {
	if len(handlers) == 0 {
		return nil, nil
	}
	var list []model.JobInstance
	err := s.DB.WithContext(ctx).
		Where("handler IN ?", handlers).
		Where("status = ? OR (status = ? AND worker_id = '')", model.InstQueued, model.InstPending).
		Order("id asc").
		Find(&list).Error
	return list, err
}

func (s *Stores) MarkStaleWorkersOffline(ctx context.Context, before time.Time) (int64, error) {
	result := s.DB.WithContext(ctx).Model(&model.WorkerNode{}).
		Where("status = ? AND (last_heartbeat_at IS NULL OR last_heartbeat_at < ?)", model.WorkerOnline, before).
		Update("status", model.WorkerOffline)
	return result.RowsAffected, result.Error
}

func (s *Stores) ListOnlineWorkersByHandler(ctx context.Context, handler string) ([]model.WorkerNode, error) {
	var all []model.WorkerNode
	if err := s.DB.WithContext(ctx).Where("status = ?", model.WorkerOnline).Find(&all).Error; err != nil {
		return nil, err
	}
	var matched []model.WorkerNode
	for _, w := range all {
		if workerSupportsHandler(w.Handlers, handler) {
			matched = append(matched, w)
		}
	}
	return matched, nil
}

func workerSupportsHandler(handlersJSON, handler string) bool {
	if handlersJSON == "" {
		return false
	}
	return containsHandler(handlersJSON, handler)
}

func containsHandler(raw, handler string) bool {
	// simple substring match for JSON array like ["stat.sync"]
	return len(handler) > 0 && (raw == handler ||
		len(raw) > 2 && (containsQuoted(raw, handler)))
}

func containsQuoted(s, h string) bool {
	quoted := `"` + h + `"`
	for i := 0; i+len(quoted) <= len(s); i++ {
		if s[i:i+len(quoted)] == quoted {
			return true
		}
	}
	return false
}
