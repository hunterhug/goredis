package redislock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// Lock 加锁
func (lock *RedisLock) Lock(ctx context.Context) error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	tracer := otel.Tracer("redis.lock")
	ctx, span := tracer.Start(ctx, "redis-lock-Lock")
	span.SetAttributes(
		attribute.String("key", lock.key),
		attribute.String("token", lock.token),
		attribute.Int64("lockTimeout", lock.lockTimeout.Milliseconds()),
		attribute.Bool("isAutoRenew", lock.isAutoRenew),
	)
	defer span.End()

	result, err := lock.Cli.Eval(ctx, lockScript, []string{lock.key}, lock.token, lock.lockTimeout.Seconds()).Result()
	if err != nil {
		span.RecordError(err) // 记录错误到 Span
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	if result != nil && result == "OK" {
		span.SetAttributes(attribute.String("result", result.(string)))
	} else {
		span.SetAttributes(attribute.String("result", fmt.Sprintf("lock acquisition failed: %v", result)))
		return errors.New(fmt.Sprintf("lock acquisition failed: %v", result))
	}

	if lock.isAutoRenew {
		lock.autoRenewCtx, lock.autoRenewCancel = context.WithCancel(lock.Context)
		go lock.autoRenew(ctx)
	}

	return nil
}

// UnLock 解锁
func (lock *RedisLock) UnLock(ctx context.Context) error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	tracer := otel.Tracer("redis.lock")
	ctx, span := tracer.Start(ctx, "redis-lock-UnLock")
	defer span.End()

	// 如果已经创建了取消函数，则执行取消操作
	if lock.autoRenewCancel != nil {
		lock.autoRenewCancel()
	}

	result, err := lock.Cli.Eval(ctx, unLockScript, []string{lock.key}, lock.token).Result()
	if err != nil {
		span.RecordError(err) // 记录错误到 Span
		return fmt.Errorf("ailed to release lock: %w", err)
	}

	if result != nil && result == "OK" {
		span.SetAttributes(attribute.String("result", result.(string)))
	} else {
		span.SetAttributes(attribute.String("result", fmt.Sprintf("lock release failed: %v", result)))
		return errors.New(fmt.Sprintf("lock release failed: %v", result))
	}

	return nil
}

// SpinLock 自旋锁
func (lock *RedisLock) SpinLock(ctx context.Context, timeout time.Duration) error {
	exp := time.Now().Add(timeout)
	tracer := otel.Tracer("redis.lock")
	ctx, span := tracer.Start(ctx, "redis-lock-SpinLock")
	span.SetAttributes(
		attribute.String("key", lock.key),
		attribute.String("token", lock.token),
		attribute.Int64("timeout", timeout.Milliseconds()),
	)
	defer span.End()

	for {
		if time.Now().After(exp) {
			span.SetAttributes(attribute.String("result", "spin lock timeout"))
			return errors.New("spin lock timeout")
		}

		// 加锁成功直接返回
		err := lock.Lock(ctx)
		if err == nil {
			return nil
		}

		// 如果加锁失败，则休眠一段时间再尝试
		select {
		case <-lock.Context.Done():
			Log.DebugContext(ctx, "spin lock done, err: %v", lock.Context.Err())
			return lock.Context.Err() // 处理取消操作
		case <-time.After(100 * time.Millisecond):
			// 自旋锁可加指数回避
			Log.DebugContext(ctx, "spin lock retrying...")
			// 继续尝试下一轮加锁
		}
	}
}

// Renew 锁手动续期
func (lock *RedisLock) Renew(ctx context.Context) error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	tracer := otel.Tracer("redis.lock")
	ctx, span := tracer.Start(ctx, "redis-lock-Renew")
	span.SetAttributes(
		attribute.String("key", lock.key),
		attribute.String("token", lock.token),
		attribute.Int64("lockTimeout", lock.lockTimeout.Milliseconds()),
	)
	defer span.End()

	result, err := lock.Cli.Eval(ctx, renewScript, []string{lock.key}, lock.token, lock.lockTimeout.Seconds()).Result()
	if err != nil {
		span.RecordError(err) // 记录错误到 Span
		return fmt.Errorf("failed to renew lock: %s", err)
	}

	if result != nil && result == "OK" {
		span.SetAttributes(attribute.String("result", result.(string)))
	} else {
		span.SetAttributes(attribute.String("result", fmt.Sprintf("lock renewal failed: %v", result)))
		return errors.New(fmt.Sprintf("lock renewal failed: %v", result))
	}

	return nil
}

// 锁自动续期
func (lock *RedisLock) autoRenew(ctx context.Context) {
	timeRenewCal := lock.lockTimeout / 3
	Log.DebugContext(ctx, "autoRenew timeRenewCal: %v", timeRenewCal)
	if timeRenewCal < 200*time.Millisecond {
		timeRenewCal = 200 * time.Millisecond
	}

	ticker := time.NewTicker(timeRenewCal)

	defer ticker.Stop()

	for {
		select {
		case <-lock.autoRenewCtx.Done():
			return
		case <-ticker.C:
			Log.DebugContext(ctx, "autoRenewing...")
			err := lock.Renew(ctx)
			if err != nil {
				Log.DebugContext(ctx, "autoRenew failed: %s", err.Error())
				return
			}
		}
	}
}
