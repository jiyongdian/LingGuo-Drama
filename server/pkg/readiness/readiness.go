package readiness

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"spiritFruit/pkg/database"
	"spiritFruit/pkg/redis"
	"time"
)

type Check struct {
	Name    string `json:"name"`
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
}

func Run() (bool, []Check) {
	checks := []Check{
		checkDatabase(),
		checkRedis(),
		checkFFmpeg(),
		checkWritableDir("uploads"),
		checkWritableDir("storage/logs"),
	}

	ready := true
	for _, check := range checks {
		if !check.OK {
			ready = false
		}
	}

	return ready, checks
}

func checkDatabase() Check {
	check := Check{Name: "database"}
	if database.SQLDB == nil {
		check.Message = "database is not initialized"
		return check
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := database.SQLDB.PingContext(ctx); err != nil {
		check.Message = err.Error()
		return check
	}

	check.OK = true
	return check
}

func checkRedis() Check {
	check := Check{Name: "redis"}
	if redis.Redis == nil {
		check.Message = "redis is not initialized"
		return check
	}

	if err := redis.Redis.Ping(); err != nil {
		check.Message = err.Error()
		return check
	}

	check.OK = true
	return check
}

func checkFFmpeg() Check {
	check := Check{Name: "ffmpeg"}
	if err := exec.Command("ffmpeg", "-version").Run(); err != nil {
		check.Message = err.Error()
		return check
	}

	check.OK = true
	return check
}

func checkWritableDir(dir string) Check {
	check := Check{Name: dir}
	if err := os.MkdirAll(dir, 0755); err != nil {
		check.Message = err.Error()
		return check
	}

	file, err := os.CreateTemp(dir, ".ready-*")
	if err != nil {
		check.Message = err.Error()
		return check
	}

	name := file.Name()
	if err := file.Close(); err != nil {
		check.Message = err.Error()
		return check
	}
	if err := os.Remove(filepath.Clean(name)); err != nil {
		check.Message = err.Error()
		return check
	}

	check.OK = true
	return check
}
