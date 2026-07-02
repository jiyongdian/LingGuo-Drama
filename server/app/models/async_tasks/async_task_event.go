package async_tasks

import (
	"encoding/json"
	"spiritFruit/app/models"
	"spiritFruit/pkg/database"
)

type AsyncTaskEvent struct {
	models.BaseModel

	TaskID  uint64 `json:"task_id" gorm:"index;not null"`
	Type    string `json:"type" gorm:"size:64;index;not null"`
	Message string `json:"message" gorm:"type:text"`
	Context string `json:"context" gorm:"type:text"`

	models.CommonTimestampsField
}

func (event *AsyncTaskEvent) TableName() string {
	return "async_task_events"
}

func CreateTaskEvent(taskID uint64, eventType string, message string, context map[string]interface{}) {
	if taskID == 0 {
		return
	}

	contextBytes, _ := json.Marshal(context)
	event := AsyncTaskEvent{
		TaskID:  taskID,
		Type:    eventType,
		Message: message,
		Context: string(contextBytes),
	}
	database.DB.Create(&event)
}
