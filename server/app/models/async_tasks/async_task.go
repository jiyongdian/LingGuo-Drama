package async_tasks

import (
	"crypto/sha1"
	"fmt"
	"spiritFruit/app/models"
	"spiritFruit/pkg/database"
	"time"
)

// 任务状态常量
const (
	StatusPending    = 0 // 排队中
	StatusProcessing = 1 // 执行中
	StatusSuccess    = 2 // 成功
	StatusFailed     = 3 // 失败
	StatusCancelled  = 4 // 已取消
)

// 任务类型常量
const (
	TypeGenerateScript = "generate_script"
	TypeGenerateImage  = "generate_image"
)

const (
	StatusNamePending   = "pending"
	StatusNameRunning   = "running"
	StatusNameSucceeded = "succeeded"
	StatusNameFailed    = "failed"
	StatusNameCancelled = "cancelled"
)

const (
	EventQueued    = "queued"
	EventStarted   = "started"
	EventProgress  = "progress"
	EventSucceeded = "succeeded"
	EventFailed    = "failed"
)

type AsyncTask struct {
	models.BaseModel

	ProjectID      uint64  `json:"project_id" gorm:"index"`       // 关联的项目ID
	RelID          uint64  `json:"rel_id" gorm:"index"`           // 关联的具体业务ID (如 script_id, character_id)
	Type           string  `json:"type" gorm:"size:64;index"`     // 任务类型
	Status         int     `json:"status" gorm:"default:0;index"` // 状态
	StatusName     string  `json:"status_name" gorm:"size:32;default:'pending';index"`
	IdempotencyKey *string `json:"idempotency_key" gorm:"size:191;index"`
	AdminID        *uint64 `json:"admin_id" gorm:"index"` // 关联的管理员ID

	Payload    string `json:"payload" gorm:"type:text"`   // 请求参数快照 (JSON)
	Result     string `json:"result" gorm:"type:text"`    // 执行结果/错误信息 (JSON)
	Process    uint64 `json:"process" gorm:"default:0"`   // 执行进度
	ErrorMsg   string `json:"error_msg" gorm:"type:text"` // 失败时的错误信息
	RetryCount uint64 `json:"retry_count" gorm:"default:0"`
	Reused     bool   `json:"-" gorm:"-"`

	StartedAt  *time.Time `json:"started_at"`  // 开始执行时间
	FinishedAt *time.Time `json:"finished_at"` // 结束时间

	models.CommonTimestampsField
}

func (task *AsyncTask) TableName() string {
	return "async_tasks"
}

func (task *AsyncTask) Create() {
	task.StatusName = statusName(task.Status)
	task.ensureIdempotencyKey()

	var existing AsyncTask
	err := database.DB.
		Where("idempotency_key = ? AND status IN ?", *task.IdempotencyKey, []int{StatusPending, StatusProcessing}).
		First(&existing).
		Error
	if err == nil {
		*task = existing
		task.Reused = true
		return
	}

	database.DB.Create(&task)
	CreateTaskEvent(task.ID, EventQueued, "task queued", map[string]interface{}{
		"type": task.Type,
	})
}

func (task *AsyncTask) Save() int64 {
	result := database.DB.Save(&task)
	return result.RowsAffected
}

// MarkAsProcessing 辅助方法：更新状态
func (task *AsyncTask) MarkAsProcessing() {
	now := time.Now()
	task.Status = StatusProcessing
	task.StatusName = StatusNameRunning
	task.Process = 10
	task.StartedAt = &now
	database.DB.Model(task).Select("Status", "StatusName", "Process", "StartedAt").Updates(task)
	CreateTaskEvent(task.ID, EventStarted, "task started", nil)
}

// UpdateProgress 单独更新进度
// 注意：不要太过频繁调用数据库，建议间隔更新（如 20%, 50%, 80%）
func (task *AsyncTask) UpdateProgress(percent uint64) {
	if percent > 100 {
		percent = 100
	}
	task.Process = percent
	database.DB.Model(task).Update("process", percent)
	CreateTaskEvent(task.ID, EventProgress, "task progress updated", map[string]interface{}{
		"process": percent,
	})
}

func (task *AsyncTask) MarkAsSuccess(result string) {
	now := time.Now()
	task.Status = StatusSuccess
	task.StatusName = StatusNameSucceeded
	task.Process = 100 // 确保完成时是 100%
	task.Result = result
	task.FinishedAt = &now
	database.DB.Save(task)
	CreateTaskEvent(task.ID, EventSucceeded, "task succeeded", nil)
}

func (task *AsyncTask) MarkAsFailed(err error) {
	now := time.Now()
	task.Status = StatusFailed
	task.StatusName = StatusNameFailed
	task.ErrorMsg = err.Error()
	task.FinishedAt = &now
	task.Save()
	CreateTaskEvent(task.ID, EventFailed, err.Error(), nil)
}

func (task *AsyncTask) MarkAsCancelled(reason string) {
	now := time.Now()
	task.Status = StatusCancelled
	task.StatusName = StatusNameCancelled
	task.ErrorMsg = reason
	task.FinishedAt = &now
	task.Save()
	CreateTaskEvent(task.ID, StatusNameCancelled, reason, nil)
}

func statusName(status int) string {
	switch status {
	case StatusProcessing:
		return StatusNameRunning
	case StatusSuccess:
		return StatusNameSucceeded
	case StatusFailed:
		return StatusNameFailed
	case StatusCancelled:
		return StatusNameCancelled
	default:
		return StatusNamePending
	}
}

func (task *AsyncTask) ensureIdempotencyKey() {
	if task.IdempotencyKey != nil && *task.IdempotencyKey != "" {
		return
	}

	sum := sha1.Sum([]byte(fmt.Sprintf("%s:%d:%d:%s", task.Type, task.ProjectID, task.RelID, task.Payload)))
	key := fmt.Sprintf("%x", sum)
	task.IdempotencyKey = &key
}
