package timerlib

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hunterhug/golog/v2"
	"github.com/redis/go-redis/v9"
)

var (
	Log = golog.New()
)

type Timer struct {
	Cli          redis.Cmdable
	service      string
	instance     string
	lockTTL      time.Duration
	task         func()
	isRunning    bool
	ctx          context.Context
	cancel       context.CancelFunc
	taskInterval time.Duration
	lock         sync.Mutex
	hasBeenRun   bool
}

func New(redisClient redis.UniversalClient, service string, lockTTL time.Duration, taskInterval time.Duration, task func()) *Timer {
	ctx, cancel := context.WithCancel(context.Background())
	if taskInterval <= 0 {
		taskInterval = 2 * time.Second
	}

	if lockTTL <= 0 {
		lockTTL = 3 * time.Second
	}

	if service == "" {
		service = "default"
	}

	return &Timer{
		Cli:          redisClient,
		service:      service,
		taskInterval: taskInterval,
		instance:     fmt.Sprintf("%s-%d", service, time.Now().UnixNano()),
		lockTTL:      lockTTL,
		task:         task,
		ctx:          ctx,
		cancel:       cancel,
		hasBeenRun:   false,
		lock:         sync.Mutex{},
	}
}

func (t *Timer) Run() error {
	t.lock.Lock()
	if t.hasBeenRun {
		t.lock.Unlock()
		return fmt.Errorf("timer has been run")
	}
	t.hasBeenRun = true
	t.lock.Unlock()

	client := t.Cli
	lockKey := fmt.Sprintf("timer:lock:%s", t.service)

	// 立即尝试获取锁
	if ok, _ := client.SetNX(t.ctx, lockKey, t.instance, t.lockTTL).Result(); ok {
		Log.Debugf("get lock: %s", lockKey)
		t.isRunning = true
		go t.runTask()
	}

	heartbeat := time.NewTicker(t.lockTTL / 3)
	defer heartbeat.Stop()

	for {
		select {
		case <-t.ctx.Done():
			return nil
		case <-heartbeat.C:
			if t.isRunning {
				// 续期
				if current, _ := client.Get(t.ctx, lockKey).Result(); current != t.instance {
					Log.Debugf("lock: %s change isRunning=false", lockKey)
					t.isRunning = false
				} else {
					Log.Debugf("lock: %s refresh ttl", lockKey)
					client.Expire(t.ctx, lockKey, t.lockTTL)
				}
			} else {
				// 尝试获取锁
				if ok, _ := client.SetNX(t.ctx, lockKey, t.instance, t.lockTTL).Result(); ok {
					Log.Debugf("get lock: %s again", lockKey)
					t.isRunning = true
					go t.runTask()
				}
			}
		}
	}
}

func (t *Timer) runTask() {
	ticker := time.NewTicker(t.taskInterval)
	defer ticker.Stop()

	for t.isRunning {
		select {
		case <-t.ctx.Done():
			return
		case <-ticker.C:
			if t.task != nil {
				t.task()
			}
			ticker.Reset(t.taskInterval)
		}
	}
}

func (t *Timer) Stop() {
	t.cancel()
	t.isRunning = false
}
