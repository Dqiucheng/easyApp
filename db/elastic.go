package db

import (
	"context"
	"easyApp/config"
	"easyApp/logger"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/mitchellh/mapstructure"
	"github.com/olivere/elastic/v7"
	"go.uber.org/zap"
)

type ElasticLogger struct {
}
type ElasticErrLogger struct {
}
type ElasticDecode struct {
}

func (w ElasticLogger) Printf(format string, v ...interface{}) {
	logger.SysL().Info("elasticsearch日志", zap.String("logData", fmt.Sprintf(format, v)))
}
func (w ElasticErrLogger) Printf(format string, v ...interface{}) {
	msg := fmt.Errorf(format, v)
	logger.SysL().Error("elasticsearch日志", zap.Error(msg))
	logger.ErrPush(context.Background(), msg)
}
func (w ElasticDecode) Decode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

var elasticsearchClientConfigs map[string]*elastic.Client

type elasticsearchClientConfig struct {
	Addresses []string // A list of Elasticsearch nodes to use.
	Username  string   // Username for HTTP Basic Authentication.
	Password  string   // Password for HTTP Basic Authentication.
}

func ConnectElasticsearch() {
	elasticsearchClientConfigs = make(map[string]*elastic.Client)
	elasticsearchConfig := config.Database.Elasticsearch

	for k, db := range elasticsearchConfig {
		var dbConf elasticsearchClientConfig
		if err := mapstructure.Decode(db, &dbConf); err != nil {
			logger.SysLog(nil).Panic("elasticsearch日志：" + config.AppMode() + ".elasticsearch[请检查配置是否正确]，err：" + err.Error())
		}

		clientOptionFunc := make([]elastic.ClientOptionFunc, 0)
		clientOptionFunc = append(clientOptionFunc, elastic.SetSniff(false)) // SetSniff启用或禁用集群嗅探器（默认情况下启用）。
		clientOptionFunc = append(clientOptionFunc, elastic.SetDecoder(ElasticDecode{}))

		clientOptionFunc = append(clientOptionFunc, elastic.SetErrorLog(ElasticErrLogger{}))
		clientOptionFunc = append(clientOptionFunc, elastic.SetInfoLog(ElasticLogger{}))
		if config.AppMode() != "release" {
			clientOptionFunc = append(clientOptionFunc, elastic.SetTraceLog(ElasticLogger{}))
		}

		clientOptionFunc = append(clientOptionFunc, elastic.SetURL(dbConf.Addresses...))
		if dbConf.Username != "" && dbConf.Password != "" {
			clientOptionFunc = append(clientOptionFunc, elastic.SetBasicAuth(dbConf.Username, dbConf.Password))
		}

		var err error
		elasticsearchClientConfigs[k], err = elastic.NewClient(clientOptionFunc...)
		if err != nil {
			logger.SysLog(nil).Panic("elasticsearch日志：" + "链接错误：" + err.Error())
		}
	}

	if len(elasticsearchConfig) > 0 {
		logger.SysLog("elasticsearchClient注册成功").Info("elasticsearch日志")
	}
}

func Es() *elastic.Client {
	return elasticsearchClientConfigs["default"]
}

// EsKey 获取指定数据链接(切换数据库)
func EsKey(key string) *elastic.Client {
	return elasticsearchClientConfigs[key]
}
