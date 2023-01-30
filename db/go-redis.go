package db

import (
	"context"
	"easyApp/config"
	"easyApp/logger"
	"errors"
	"github.com/mitchellh/mapstructure"
	redis "github.com/redis/go-redis/v9"
	"reflect"
	"time"
)

var redisPool map[string]*redis.Client

type rdConf struct {
	Host         string // 主机
	Port         string // 端口
	Password     string // 密码
	DB           int    // 数据库
	MinIdleConns int    // 最小空闲连接数，在启动阶段创建指定数量的Idle连接
	MaxIdleConns int    // 最大空闲链接数
}

func ConnectRedisS() {
	redisPool = make(map[string]*redis.Client) // 初始化redis

	dbConfig := config.Database.Redis
	for k, db := range dbConfig {
		var dbConf rdConf
		if err := mapstructure.Decode(db, &dbConf); err != nil {
			logger.SysLog(nil).Panic("Redis日志：" + config.AppMode() + ".redis[请检查配置是否正确]，err：" + err.Error())
		}
		onnectRedis(k, dbConf)
	}
}

func onnectRedis(dbName string, rdConf rdConf) {
	t := reflect.TypeOf(rdConf)
	v := reflect.ValueOf(rdConf)
	for k := 0; k < t.NumField(); k++ {
		if v.Field(k).String() == "" && t.Field(k).Name != "Password" {
			if t.Field(k).Name == "Host" {
				return
			}
			logger.SysLog(nil).Panic("Redis日志：" + config.AppMode() + ".redis." + t.Field(k).Name + " 不能为空")
		}
	}

	redisPool[dbName] = redis.NewClient(&redis.Options{
		// 连接信息
		Network:  "tcp",                           // 网络类型，tcp or unix，默认tcp
		Addr:     rdConf.Host + ":" + rdConf.Port, // 主机名+冒号+端口，默认localhost:6379
		Password: rdConf.Password,                 // 密码
		DB:       rdConf.DB,                       // redis数据库index

		//连接池容量及闲置连接数量
		PoolFIFO: false, // 连接池类型。true为FIFO池，false为LIFO池。fifo比lifo有更高的开销。
		//PoolSize:     15,    // 连接池最大socket连接数，默认为4倍CPU数， 4 * runtime.NumCPU
		MinIdleConns: rdConf.MinIdleConns, // 在启动阶段创建指定数量的Idle连接，并长期维持idle状态的连接数不少于指定数量。
		MaxIdleConns: rdConf.MaxIdleConns, // 最大空闲链接数

		// 超时
		DialTimeout:  5 * time.Second, // 连接建立超时时间，默认5秒。
		ReadTimeout:  3 * time.Second, // 读超时，默认3秒， -1表示取消读超时
		WriteTimeout: 3 * time.Second, // 写超时，默认等于读超时
		PoolTimeout:  4 * time.Second, // 当所有连接都处在繁忙状态时，客户端等待可用连接的最大等待时长，默认为读超时+1秒。

		// 命令执行失败时的重试策略
		MaxRetries:      0,                      // 命令执行失败时，最多重试多少次，默认为0即不重试
		MinRetryBackoff: 8 * time.Millisecond,   // 每次计算重试间隔时间的下限，默认8毫秒，-1表示取消间隔
		MaxRetryBackoff: 512 * time.Millisecond, // 每次计算重试间隔时间的上限，默认512毫秒，-1表示取消间隔
	})
	rd, err := RdKey(dbName)
	if err != nil {
		logger.SysLog(nil).Panic("Redis日志：" + "获取Redis链接错误：" + err.Error())
	}

	if err := rd.Ping(context.Background()).Err(); err != nil {
		logger.SysLog(nil).Panic("Redis日志：" + "redis链接错误：" + err.Error())
	}

	logger.SysLog(dbName + " Redis连接池注册成功").Info("Redis日志")
}

func Rd() *redis.Client {
	return redisPool["default"]
}

// RdKey 获取指定数据链接(切换数据库)
func RdKey(key string) (*redis.Client, error) {
	if o, ok := redisPool[key]; ok {
		return o, nil
	}
	return nil, errors.New("未获取到【" + key + "】RedisDB")
}
