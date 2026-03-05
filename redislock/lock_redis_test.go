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
	ctx := context.Background()
	redisClient, err := goredis.New(nil)
	if err != nil {
		golog.ErrorContext(ctx, "NewRedisUrl err: %s", err.Error())
		return
	}
	
	lock := New(
		ctx, redisClient,
		WithAutoRenew(),
		WithKeyPrefix("/locks/xx"),
		WithKey(fmt.Sprintf("package-xx")),
		WithTimeout(time.Duration(50)*time.Second))
	err = lock.Lock(ctx)
	if err != nil {
		golog.ErrorContext(ctx, "Lock err: %s", err.Error())
		return
	}

	time.Sleep(time.Duration(100) * time.Second)
	defer func(lock LockInter) {
		err := lock.UnLock(ctx)
		if err != nil {
			golog.ErrorContext(ctx, "UnLock err: %s", err.Error())
		}
	}(lock)

	time.Sleep(time.Duration(5) * time.Second)
	golog.InfoContext(ctx, "Lock success")
}

func TestLock2(t *testing.T) {
	ctx := context.Background()
	redisClient, err := goredis.New(nil)
	if err != nil {
		golog.ErrorContext(ctx, "NewRedisUrl err: %s", err.Error())
		return
	}
	lock := New(
		ctx, redisClient,
		WithAutoRenew(),
		WithKeyPrefix("/locks/xx"),
		WithKey(fmt.Sprintf("package-xx")),
		WithTimeout(time.Duration(50)*time.Second))
	err = lock.Lock(ctx)
	if err != nil {
		golog.ErrorContext(ctx, "Lock err: %s", err.Error())
		return
	}

	time.Sleep(time.Duration(100) * time.Second)
	defer func(lock LockInter) {
		err := lock.UnLock(ctx)
		if err != nil {
			golog.ErrorContext(ctx, "UnLock err: %s", err.Error())
		}
	}(lock)

	time.Sleep(time.Duration(5) * time.Second)
	golog.InfoContext(ctx, "Lock success")
}
