package v1

import (
	"github.com/gin-gonic/gin"
	"spiritFruit/app/models/async_tasks"
	"spiritFruit/pkg/database"
	"spiritFruit/pkg/response"
)

type TasksController struct {
	BaseADMINController
}

// Show 获取任务详情 (用于轮询)
func (ctrl *TasksController) Show(c *gin.Context) {
	id := c.Param("id")
	var task async_tasks.AsyncTask

	// 查询任务表
	if err := database.DB.First(&task, id).Error; err != nil {
		response.Abort404(c, "任务不存在")
		return
	}

	var events []async_tasks.AsyncTaskEvent
	database.DB.
		Where("task_id = ?", task.ID).
		Order("id desc").
		Limit(20).
		Find(&events)

	// 返回包含 Process(进度) 和 Status(状态) 的数据
	response.Data(c, gin.H{
		"task":   task,
		"events": events,
	})
}

func (ctrl *TasksController) Cancel(c *gin.Context) {
	id := c.Param("id")
	var task async_tasks.AsyncTask

	if err := database.DB.First(&task, id).Error; err != nil {
		response.Abort404(c, "任务不存在")
		return
	}

	if task.Status == async_tasks.StatusSuccess || task.Status == async_tasks.StatusFailed || task.Status == async_tasks.StatusCancelled {
		response.Abort400(c, "任务已结束，不能取消")
		return
	}

	task.MarkAsCancelled("cancelled by user")
	response.Data(c, task)
}
