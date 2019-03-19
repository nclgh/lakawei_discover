package lakawei_discover

import (
	"sync"
	"time"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/nclgh/lakawei_discover/config"
	"github.com/nclgh/lakawei_discover/helper"
)

const (
	RedisHeartbeatInterval = 100 * time.Millisecond
	RetryFindRedisMaster   = 10
)

var (
	discoverRedisClient *redis.Client
	masterAddr          string

	mlock sync.Mutex
)

func GetRedisClient() *redis.Client {
	mlock.Lock()
	defer mlock.Unlock()
	return discoverRedisClient
}

func SetRedisClient(cli *redis.Client, addr string) {
	mlock.Lock()
	defer mlock.Unlock()
	discoverRedisClient = cli
	masterAddr = addr
	logrus.Infof("redis master switch to %v", addr)
}

var (
	ErrRedisMasterNotFound = errors.New("redis master not found")
)

func initRedisClient() {
	err := findRedisMaster()
	if err != nil {
		panic(err)
	}
	go ensureRedisMaster()
}

func ensureRedisMaster() {
	defer helper.RecoverPanic(func(err interface{}, stacks string) {
		logrus.Errorf("ensureRedisMaster panic: %v, stack: %v", err, stacks)
		time.Sleep(5 * time.Second)
		ensureRedisMaster()
	})
	for {
		err := GetRedisClient().Ping().Err()
		if err != nil {
			logrus.Errorf("ping redis cli failed. addr: %v, err: %v", masterAddr, err)
			for i := 0; i < RetryFindRedisMaster; i++ {
				err = findRedisMaster()
				if err == nil {
					break
				}
				if i == RetryFindRedisMaster-1 {
					logrus.Errorf("redis cluster unstable now")
				}
				time.Sleep(RedisHeartbeatInterval)
			}
		}
		time.Sleep(RedisHeartbeatInterval)
	}
}

func findRedisMaster() error {
	c := config.GetConfig()
	for _, addr := range c.DiscoverRedisList {
		cli := redis.NewClient(&redis.Options{
			Addr: addr,
		})
		err := cli.Set("is_master", 1, time.Second).Err()
		if err != nil {
			//logrus.Infof("connect to redis but slave addr: %v", addr)
			continue
		}
		SetRedisClient(cli, addr)
		return nil
	}
	return ErrRedisMasterNotFound
}
