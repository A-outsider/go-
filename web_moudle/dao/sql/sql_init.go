package sql

import (
	"fmt"
	"github.com/go-redis/redis"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"new/web_moudle/settings"
)

var DB *gorm.DB
var RE *redis.Client

func initMysql(conf *settings.MySqlConfig) {

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.User,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.DbName)
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		zap.L().Error("connect DB failed", zap.Error(err))
		panic(err)
	}

}

func initRedis(conf *settings.RedisConfig) {

	RE = redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.DB,
		PoolSize: conf.PoolSize,
	})

	_, err := RE.Ping().Result()
	if err != nil {
		zap.L().Error("connect RE failed", zap.Error(err))
		panic(err)
	}
}

func Init() {
	initMysql(settings.Conf.MySqlConfig)
	initRedis(settings.Conf.RedisConfig)
}
