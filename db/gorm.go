package db

import (
	"easyApp/config"
	"easyApp/logger"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var orm map[string]*gorm.DB

type msConf struct {
	Host         string // 主机
	Port         string // 端口
	Username     string // 用户名
	Password     string // 密码
	Database     string // 数据库名称
	Prefix       string // 数据库表前缀
	MaxIdleConns int    // 最大空闲连接数
	MaxOpenConns int    // 最大打开的连接数
	Charset      string // 数据库编码
}

// Writer 自定义日志
type Writer struct {
}

func (w Writer) Printf(format string, args ...interface{}) {
	logger.SysLog(fmt.Sprintf(format, args...)).Info("MySQL日志")
}

func ConnectMySQLS() {
	orm = make(map[string]*gorm.DB) // 初始化orm

	dbConfig := config.Database.MySQL
	for k, db := range dbConfig {
		var dbConf msConf
		if err := mapstructure.Decode(db, &dbConf); err != nil {
			logger.SysLog(nil).Panic("MySQL日志：" + config.AppMode() + ".mysql[请检查配置是否正确]，err：" + err.Error())
		}
		connectMySQL(k, dbConf)
	}
}

func connectMySQL(dbName string, dbConf msConf) {
	t := reflect.TypeOf(dbConf)
	v := reflect.ValueOf(dbConf)
	for k := 0; k < t.NumField(); k++ {
		if v.Field(k).String() == "" && t.Field(k).Name != "Prefix" {
			if t.Field(k).Name == "Host" {
				return
			}
			logger.SysLog(nil).Panic("MySQL日志：" + config.AppMode() + ".mysql." + t.Field(k).Name + " 不能为空")
		}
	}

	dsn := strings.Builder{}
	dsn.WriteString(dbConf.Username + ":" + dbConf.Password)
	dsn.WriteString("@tcp(")
	dsn.WriteString(dbConf.Host + ":" + dbConf.Port)
	dsn.WriteString(")/")
	dsn.WriteString(dbConf.Database)
	dsn.WriteString("?charset=" + dbConf.Charset)
	dsn.WriteString("&parseTime=true&loc=Local")

	var LogLevel gormLogger.LogLevel
	if config.AppMode() == "release" {
		LogLevel = gormLogger.Warn
	} else {
		LogLevel = gormLogger.Info
	}

	newLogger := gormLogger.New(
		Writer{},
		gormLogger.Config{
			SlowThreshold:             2 * time.Second, // 慢 SQL 阈值
			LogLevel:                  LogLevel,        // 日志级别，ORM 定义了这些日志级别：Silent、Error、Warn、Info
			IgnoreRecordNotFoundError: true,            // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  false,           // 禁用彩色打印
		},
	)
	var err error
	orm[dbName], err = gorm.Open(mysql.Open(dsn.String()), &gorm.Config{
		Logger:      newLogger,
		PrepareStmt: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   dbConf.Prefix, // 表名前缀，`User` 的表名应该是 `t_users`
			SingularTable: true,          // 使用单数表名，启用该选项，此时，`User` 的表名应该是 `t_user`
		},
	})
	if err != nil {
		logger.SysLog(nil).Panic("MySQL日志：" + "Fail to open mysql：" + err.Error())
	}

	sqlDB, errDB := orm[dbName].DB()
	if errDB != nil {
		logger.SysLog(nil).Panic("MySQL日志：" + "获取*sql.DB失败：" + err.Error())
	}
	sqlDB.SetMaxIdleConns(dbConf.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbConf.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	logger.SysLog(dbName + " MySQL连接池注册成功").Info("MySQL日志")
}

// Ms 获取默认数据链接
func Ms() *gorm.DB {
	return orm["default"]
}

// MsKey 获取指定数据链接(切换数据库)
func MsKey(key string) (*gorm.DB, error) {
	if o, ok := orm[key]; ok {
		return o, nil
	}
	return nil, errors.New("未获取到【" + key + "】MySQLDb")
}
