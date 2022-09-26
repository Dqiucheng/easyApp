package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"log"
)

var AppViper *viper.Viper
var AppSecretViper *viper.Viper
var DatabaseViper *viper.Viper
var CallHostViper *viper.Viper

func init() {
	AppViper = viper.New() // 一定要先注册app
	AppSecretViper = viper.New()
	DatabaseViper = viper.New()
	CallHostViper = viper.New()

	viperInit(AppViper, "app")
	viperInit(AppSecretViper, "appsecret")
	viperInit(DatabaseViper, "database")
	viperInit(CallHostViper, "callhost")

	// 解析至结构体
	viperUnmarshal()
}

// viperInit 预加载配置文件
func viperInit(vipers *viper.Viper, configName string) {
	// 设置目录
	vipers.AddConfigPath("config/tomlConfig")

	// 设置文件名
	vipers.SetConfigName(configName)

	// 设置后缀
	vipers.SetConfigType("toml")

	// 配置验证
	if err := vipers.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(fmt.Errorf("%s配置文件未找到: %w", configName, err))
		} else {
			panic(fmt.Errorf("%s配置文件异常: %w", configName, err))
		}
	}

	if AppViper.GetInt("app.IsOnConfigChange") == 1 {
		// 配置文件热加载
		vipers.OnConfigChange(func(e fsnotify.Event) {
			log.Println("配置文件有变更:", e.Name)
			// 解析至结构体
			viperUnmarshal()
		})
		vipers.WatchConfig()
	}
}

// AppMode 获取服务运行模式
func AppMode() string {
	return App.SetMode
}

// viperUnmarshal 配置文件解析至结构体
func viperUnmarshal() {
	appConfig := AppViper.GetStringMap("app")
	if err := mapstructure.Decode(appConfig, &App); err != nil {
		panic(fmt.Errorf("app配置文件解析异常: %w", err))
	}

	appSecretConfig := AppSecretViper.GetStringMap(AppMode())
	if err := mapstructure.Decode(appSecretConfig, &AppSecret); err != nil {
		panic(fmt.Errorf("appsecret配置文件解析异常: %w", err))
	}

	databaseConfig := DatabaseViper.GetStringMap(AppMode())
	if err := mapstructure.Decode(databaseConfig, &Database); err != nil {
		panic(fmt.Errorf("database配置文件解析异常: %w", err))
	}

	callhostConfig := CallHostViper.GetStringMap(AppMode())
	if err := mapstructure.Decode(callhostConfig, &CallHost); err != nil {
		panic(fmt.Errorf("database配置文件解析异常: %w", err))
	}
}
