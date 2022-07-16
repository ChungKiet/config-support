package redislock

import (
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis"
)

var rs *redsync.Redsync

func InitPool(pool redis.Pool) {
	rs = redsync.New(pool)
}

func simpleLock(mutex *redsync.Mutex) (*redsync.Mutex, error) {
	if err := mutex.Lock(); err != nil {
		return nil, err
	}

	return mutex, nil
}

func LockCustomRetry(mutexId string, retryCount int) (*redsync.Mutex, error) {
	retryOption := redsync.WithTries(retryCount)
	mutex := rs.NewMutex(mutexId, retryOption)
	return simpleLock(mutex)
}

func Unlock(mutex *redsync.Mutex) (bool, error) {
	return mutex.Unlock()
}
