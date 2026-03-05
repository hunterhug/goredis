package goredis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	*redis.Client
	config *Config
}

type Config struct {
	Addr            string
	Password        string
	DB              int
	PoolSize        int
	MaxIdleConns    int
	MinIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	PoolTimeout     time.Duration
}

func GetDefaultConfig() *Config {
	return &Config{
		Addr:            "127.0.0.1:6379",
		Password:        "",
		DB:              0,
		PoolSize:        200,
		MaxIdleConns:    40,
		MinIdleConns:    20,
		ConnMaxLifetime: 2 * time.Hour,    // 连接最大存活时间（避免长连接占用资源）
		ConnMaxIdleTime: 5 * time.Minute,  // 空闲连接超时回收
		DialTimeout:     5 * time.Second,  // 连接建立超时
		ReadTimeout:     3 * time.Second,  // 单次读操作超时
		WriteTimeout:    3 * time.Second,  // 单次写操作超时
		PoolTimeout:     15 * time.Second, // 必须添加！获取连接超时
	}
}

func New(config *Config) (*Client, error) {
	if config == nil {
		config = GetDefaultConfig()
	}
	client := redis.NewClient(&redis.Options{
		Addr:            config.Addr,     // 直接使用传入的地址
		Password:        config.Password, // 如果没有密码则留空
		DB:              config.DB,       // 默认使用 DB 0
		PoolSize:        config.PoolSize,
		MaxIdleConns:    config.MaxIdleConns,
		MinIdleConns:    config.MinIdleConns,
		ConnMaxLifetime: config.ConnMaxLifetime, // 连接最大存活时间（避免长连接占用资源）
		ConnMaxIdleTime: config.ConnMaxIdleTime, // 空闲连接超时回收
		DialTimeout:     config.DialTimeout,     // 连接建立超时
		ReadTimeout:     config.ReadTimeout,     // 单次读操作超时
		WriteTimeout:    config.WriteTimeout,    // 单次写操作超时
		PoolTimeout:     config.PoolTimeout,     // 必须添加！获取连接超时
	})

	// 检查连接是否正常
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		_ = client.Close()
		return nil, err
	}

	resultClient := &Client{
		Client: client,
		config: config,
	}

	return resultClient, nil
}

func (c *Client) GetRedisClient() *redis.Client {
	return c.Client
}

func (c *Client) Close() error {
	return c.Client.Close()
}
