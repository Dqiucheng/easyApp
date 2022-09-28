package config

var App appStruct
var AppSecret appSecretStruct
var Database databaseStruct
var CallHost callHostStruct

// appStruct 系统相关配置
type appStruct struct {
	AppName          string  // 项目名称
	SetMode          string  // 启动模式：release、test、debug
	HttpPort         string  // 启动端口
	SlowReqThreshold float64 // 慢请求阈值（单位秒）
	ReqBurst         int     // 请求最大令牌数(每秒会产生100个令牌)
	IsOnConfigChange int     // 配置文件是否热加载，这里不建议开启。1开启、0关闭

	CertFile string // HTTPS证书
	KeyFile  string // HTTPS证书

	DLogRootPath         string // 日志跟目录，绝对路径
	DLogSetLevel         string // 日志级别 fatal、panic、error、warn、info、debug
	DLogMaxSize          int64  // 日志文件最大大小，单位兆，0不限制大小
	DLogMaxAge           int    // 日志文件最大保存天数，如果设置，不建议设置太多，避免一次性处理的文件太多造成卡顿
	DLogMillEveryDayHour int    // 每日N点开始执行陈旧文件处理，陈旧文件含义建立在DLogMaxAge参数之上，当DLogMaxAge为0时不会触发
	DLogIsInStdout       int    // 是否同时输出到控制台。1是、0否。当SetMode为“debug”模式时同样会输出到控制台
}

// appSecretStruct 加密相关配置
type appSecretStruct struct {
	SignAccountSecret map[string]string
	AesAccountSecret  map[string]string
}

// callHostStruct 请求地址相关配置
type callHostStruct struct {
	CallHost map[string]string
}

// databaseStruct 数据库相关配置
type databaseStruct struct {
	MySQL         map[string]map[string]interface{}
	Redis         map[string]map[string]interface{}
	Elasticsearch map[string]map[string]interface{}
}
