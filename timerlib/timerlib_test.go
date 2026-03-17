package timerlib

import (
	"context"
	"testing"
	"time"

	"github.com/hunterhug/golog/v2"
	"github.com/hunterhug/goredis"
)

func TestTimer(t *testing.T) {
	Log.SetLevel(golog.DebugLevel)
	Log.InitLogger()
	ctx := context.Background()
	redisClient, err := goredis.New(nil)
	if err != nil {
		Log.ErrorContext(ctx, "NewRedisUrl err: %s", err.Error())
		return
	}

	timer := New(redisClient, "xx", 2*time.Second, 2*time.Second, func() {
		Log.InfoContext(ctx, "I am timer")
	})

	go func() {
		err = timer.Run()
		if err != nil {
			Log.ErrorContext(ctx, "NewTimer err: %s", err.Error())
			return
		}
	}()

	err = timer.Run()
	if err != nil {
		Log.ErrorContext(ctx, "NewTimer err: %s", err.Error())
		return
	}
}
