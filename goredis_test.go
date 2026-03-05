package goredis

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	client, err := New(&Config{
		Addr:            "127.0.0.1:6379",
		Password:        "",
		DB:              0,
		PoolSize:        200,
		MaxIdleConns:    40,
		MinIdleConns:    20,
		ConnMaxLifetime: 2 * time.Hour,
		ConnMaxIdleTime: 5 * time.Minute,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolTimeout:     15 * time.Second,
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	ctx := context.Background()
	ex := time.Hour
	result, err := client.Set(ctx, "a", "b", ex).Result()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(result)

	result, err = client.Get(ctx, "a").Result()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(result)

	duration, err := client.TTL(ctx, "a").Result()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(duration)
}
