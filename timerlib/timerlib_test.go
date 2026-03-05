package timerlib

import (
	"context"
	"testing"
	"time"

	"github.com/hunterhug/golog/v2"
	"github.com/hunterhug/goredis"
)

func TestTimer(t *testing.T) {
	golog.SetLevel(golog.DebugLevel)
	golog.InitLogger()
	ctx := context.Background()
	redisClient, err := goredis.New(nil)
	if err != nil {
		golog.ErrorContext(ctx, "NewRedisUrl err: %s", err.Error())
		return
	}

	timer := New(redisClient, "xx", 2*time.Second, 2*time.Second, func() {
		golog.InfoContext(ctx, "I am timer")
	})

	err = timer.Run()
	if err != nil {
		golog.ErrorContext(ctx, "NewTimer err: %s", err.Error())
		return
	}

}
