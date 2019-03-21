package config

import (
	"sync"
	"github.com/jinzhu/configor"
)

type c struct {
	DiscoverRedisList []string `yaml:"discoverRedisList"`
}

var config *c

func initConfig() {
	config = &c{}
	configor.New(&configor.Config{}).Load(config, "/opt/lakawei/conf/discover/config.yml")
}

var once sync.Once

func GetConfig() *c {
	once.Do(func() { initConfig() })
	return config
}
