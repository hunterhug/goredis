package redislock

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hunterhug/golog/v2"
	"github.com/hunterhug/goredis"
)

func TestLock(t *testing.T) {
	Log.SetLevel(golog.DebugLevel)
	Log.InitLogger()
	ctx := context.Background()
	redisClient, err := goredis.New(nil)
	if err != nil {
		Log.ErrorContext(ctx, "NewRedisUrl err: %s", err.Error())
		return
	}

	lock := New(
		ctx, redisClient,
		WithAutoRenew(),
		WithKeyPrefix("/locks/xx2"),
		WithKey(fmt.Sprintf("package-xx")),
		WithTimeout(time.Duration(5)*time.Second))
	err = lock.Lock(ctx)
	if err != nil {
		Log.ErrorContext(ctx, "Lock err: %s", err.Error())
		return
	}

	time.Sleep(time.Duration(10) * time.Second)
	defer func(lock LockInter) {
		err := lock.UnLock(ctx)
		if err != nil {
			Log.ErrorContext(ctx, "UnLock err: %s", err.Error())
		}
	}(lock)

	time.Sleep(time.Duration(5) * time.Second)
	Log.InfoContext(ctx, "Lock success")
}

func TestLock2(t *testing.T) {
	Log.SetLevel(golog.DebugLevel)
	Log.InitLogger()
	ctx := context.Background()
	redisClient, err := goredis.New(nil)
	if err != nil {
		Log.ErrorContext(ctx, "NewRedisUrl err: %s", err.Error())
		return
	}
	lock := New(
		ctx, redisClient,
		WithAutoRenew(),
		WithKeyPrefix("/locks/xx2"),
		WithKey(fmt.Sprintf("package-xx")),
		WithTimeout(time.Duration(50)*time.Second))
	err = lock.SpinLock(ctx, 10*time.Second)
	if err != nil {
		Log.ErrorContext(ctx, "Lock err: %s", err.Error())
		return
	}

	time.Sleep(time.Duration(100) * time.Second)
	defer func(lock LockInter) {
		err := lock.UnLock(ctx)
		if err != nil {
			Log.ErrorContext(ctx, "UnLock err: %s", err.Error())
		}
	}(lock)

	time.Sleep(time.Duration(5) * time.Second)
	Log.InfoContext(ctx, "Lock success")
}
