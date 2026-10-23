package redis_test

import (
	"testing"
	"time"

	"code-base-golang/internal/adapters/redis"
	"code-base-golang/internal/config"
)

func TestRedis_New_Unreachable(t *testing.T) {
	cfg := config.Config{
		RedisHost:        "127.0.0.1",
		RedisPort:        "59999", // non-existent port
		RedisDialTimeout: 50 * time.Millisecond,
	}

	client, err := redis.New(cfg)
	if err == nil {
		t.Fatal("expected error when connecting to unreachable redis, got nil")
	}
	if client != nil {
		t.Fatal("expected client to be nil on failed connection")
	}
}
