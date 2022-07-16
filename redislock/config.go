package redislock

type RedisConnectionConfig struct {
	Addr             string
	UserName         string
	Password         string
	MasterName       string
	SentinelAddrs    []string
	SentinelPassword string
}
