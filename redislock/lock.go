package redislock

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

func init() {
	Log.InitLogger()
}

type LockInter interface {
	// Lock 加锁
	Lock(ctx context.Context) error

	// UnLock 解锁
	UnLock(ctx context.Context) error

	// SpinLock 自旋锁
	SpinLock(ctx context.Context, timeout time.Duration) error

	// Renew 手动续期
	Renew(ctx context.Context) error
}

type RedisLock struct {
	context.Context
	Cli             redis.Cmdable
	keyPrefix       string
	key             string
	token           string
	lockTimeout     time.Duration
	isAutoRenew     bool
	autoRenewCtx    context.Context
	autoRenewCancel context.CancelFunc
	mutex           sync.Mutex
}

// 默认锁超时时间
const lockTime = 5 * time.Second

type Option func(lock *RedisLock)

func New(ctx context.Context, redisClient redis.UniversalClient, options ...Option) LockInter {
	lock := &RedisLock{
		Context:     ctx,
		Cli:         redisClient,
		lockTimeout: lockTime,
	}
	for _, f := range options {
		f(lock)
	}

	if lock.keyPrefix == "" {
		lock.keyPrefix = "/locks"
	}

	lock.key = fmt.Sprintf("%s/%s", lock.keyPrefix, lock.key)

	// token 自动生成
	if lock.token == "" {
		lock.token = fmt.Sprintf("token_%d", time.Now().UnixNano())
	}

	return lock
}

// WithKey 设置锁的key
func WithKey(key string) Option {
	return func(lock *RedisLock) {
		lock.key = key
	}
}

// WithKeyPrefix 设置锁前缀
func WithKeyPrefix(keyPrefix string) Option {
	return func(lock *RedisLock) {
		lock.keyPrefix = keyPrefix
	}
}

// WithTimeout 设置锁过期时间
func WithTimeout(timeout time.Duration) Option {
	return func(lock *RedisLock) {
		lock.lockTimeout = timeout
	}
}

// WithAutoRenew 是否开启自动续期
func WithAutoRenew() Option {
	return func(lock *RedisLock) {
		lock.isAutoRenew = true
	}
}

// WithToken 设置锁的Token
func WithToken(token string) Option {
	return func(lock *RedisLock) {
		lock.token = token
	}
}
