package settings

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var Conf = new(Config)

type Config struct {
	*AppConfig   `mapstructure:"app"`
	*LogConfig   `mapstructure:"log"`
	*MySqlConfig `mapstructure:"mysql"`
	*RedisConfig `mapstructure:"redis"`
}

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Mode    string `mapstructure:"mode"`
	Version string `mapstructure:"version"`
	Post    string `mapstructure:"port"`
	Host    string `mapstructure:"host"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

type MySqlConfig struct {
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DbName       string `mapstructure:"dbname"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"DB"`
	PoolSize int    `mapstructure:"poolSize"`
}

func ViperInit() (err error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml") //一般用于远程中心的
	viper.AddConfigPath("./settings")
	err = viper.ReadInConfig() // 读取配置信息
	if err != nil {
		panic(err)
	}
	if err := viper.Unmarshal(Conf); err != nil {
		fmt.Println("viper.Unmarshal failed : ", err)
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("配置文件已修改")
		if err := viper.Unmarshal(Conf); err != nil {
			panic("viper.Unmarshal failed : ")
		}
	})

	return
}
