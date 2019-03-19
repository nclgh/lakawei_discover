package lakawei_discover

import (
	"math"
	"time"
	"sync"
	"strconv"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"github.com/nclgh/lakawei_discover/helper"
)

const (
	ServiceHeartbeat = 500 * time.Millisecond

	AliveAddrInterval = 2
)

type Service struct {
	ServiceName string
	ServiceAddr string
}

var (
	service  *Service
	initOnce sync.Once
)

func Register(sName string, addr string) {
	initOnce.Do(initRedisClient)

	service = &Service{
		ServiceName: sName,
		ServiceAddr: addr,
	}

	err := GetRedisClient().ZAdd(sName, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: addr,
	}).Err()
	if err != nil {
		panic(err)
	}

	go service.keepHeartbeat()
}

func Unregister() {
	GetRedisClient().ZRem(service.ServiceName, service.ServiceAddr)
	logrus.Infof("exit service discover. server: %v, addr: %v", service.ServiceName, service.ServiceAddr)
}

func GetServiceAddr(sName string) ([]string) {
	initOnce.Do(initRedisClient)
	for i := 0; i < 3; i++ {
		cli := GetRedisClient()
		min := strconv.FormatInt(time.Now().Unix()-AliveAddrInterval, 10)
		results, err := cli.ZRangeByScore(sName, redis.ZRangeBy{
			Min: min, Max: "+inf", Offset: 0, Count: math.MaxInt64}).Result()
		if err != nil {
			logrus.Errorf("query redis service ip failed. err: %v", err)
			time.Sleep(time.Second)
			continue
		}
		return results
	}
	logrus.Errorf("service discover unstable")
	return nil
}

func (s *Service) keepHeartbeat() {
	defer helper.RecoverPanic(func(err interface{}, stacks string) {
		logrus.Errorf("keepHeartbeat panic: %v, stack: %v", err, stacks)
		time.Sleep(5 * time.Second)
		s.keepHeartbeat()
	})
	for {
		err := GetRedisClient().ZAdd(s.ServiceName, redis.Z{
			Score:  float64(time.Now().Unix()),
			Member: s.ServiceAddr,
		}).Err()
		if err != nil {
			logrus.Errorf("heartbeat err: %", err)
		}
		time.Sleep(ServiceHeartbeat)
	}
}
