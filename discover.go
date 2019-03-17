package lakawei_discover

import (
	"time"
	"github.com/sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/nclgh/lakawei_discover/helper"
)

type Service struct {
	ServiceToken string
	ServiceAddr  string
}

var (
	service *Service
)

func Register(sToken string, addr string) {
	initRedisClient()

	service = &Service{
		ServiceToken: sToken,
		ServiceAddr:  addr,
	}

	err := GetRedisClient().ZAdd(sToken, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: addr,
	}).Err()
	if err != nil {
		panic(err)
	}

	go service.keepHeartbeat()
}

func (s *Service) Unregister() {
	GetRedisClient().ZRem(s.ServiceToken, s.ServiceAddr)
	logrus.Infof("exit service discover. server: %v, addr: %v", s.ServiceToken, s.ServiceAddr)
}

func (s *Service) keepHeartbeat() {
	defer helper.RecoverPanic(func(err interface{}, stacks string) {
		logrus.Errorf("keepHeartbeat panic: %v, stack: %v", err, stacks)
		time.Sleep(5 * time.Second)
		s.keepHeartbeat()
	})
	for {
		err := GetRedisClient().ZAdd(s.ServiceToken, redis.Z{
			Score:  float64(time.Now().Unix()),
			Member: s.ServiceAddr,
		}).Err()
		if err != nil {
			logrus.Errorf("heartbeat err: %", err)
		}
		time.Sleep(5 * time.Second)
	}
}
