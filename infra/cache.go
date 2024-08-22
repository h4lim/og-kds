package infra

import (
	"github.com/patrickmn/go-cache"
	"time"
)

var (
	Cache            *cache.Cache
	GlobalCacheModel CacheModel
)

type CacheModel struct {
	Expired int64
	Purge   int64
}

type CacheModelContext struct {
	cacheModel CacheModel
}

type ICacheConfig interface {
	Setup() *error
}

func NewCacheConfig(cacheModel CacheModel) ICacheConfig {
	return CacheModelContext{
		cacheModel: cacheModel,
	}
}
func (cm CacheModelContext) Setup() *error {

	expiredTime := time.Minute * time.Duration(cm.cacheModel.Expired)
	purgeTime := time.Minute * time.Duration(cm.cacheModel.Purge)
	Cache = cache.New(expiredTime, purgeTime)

	GlobalCacheModel = cm.cacheModel

	return nil
}
