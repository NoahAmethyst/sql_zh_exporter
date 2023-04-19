package config

import (
	"fmt"
	"github.com/spf13/viper"
	"sync"
)

var cache configCache

type configCache struct {
	cache map[string]string
	sync.RWMutex
}

func GetConfig(key, fileName string) string {
	// set path of yaml
	if len(cache.get(key)) > 0 {
		return cache.get(key)
	}
	viper.SetConfigFile(fileName)
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("read config from yml failed:%s", err))
	}
	value := viper.GetString(key)
	cache.add(key, value)
	// read value from yaml
	return value
}

func (c *configCache) add(key, value string) {
	c.Lock()
	defer c.Unlock()
	c.cache[key] = value
}

func (c *configCache) get(key string) string {
	c.RLock()
	defer c.RUnlock()
	return c.cache[key]
}

func init() {
	cache = configCache{
		cache:   map[string]string{},
		RWMutex: sync.RWMutex{},
	}
}
