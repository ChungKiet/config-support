package redislock

import (
	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4/redis"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
)

var client *goredislib.Client

func NewPool() redis.Pool {
	return goredis.NewPool(client)
}

func ConnectRedisV1(config *RedisConnectionConfig) {
	client = goredislib.NewClient(&goredislib.Options{
		Addr:     config.Addr,
		Username: config.UserName,
		Password: config.Password,
	})
	InitPool(NewPool())
}

func GetClient() *goredislib.Client {
	return client
}
