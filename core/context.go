package core

import (
	"bytes"
	"easyApp/config"
	"easyApp/logger"
	"easyApp/util"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

type Context struct {
	*gin.Context
}

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		// skip if tag key says it should be ignored
		if name == "-" {
			return ""
		}
		return name
	})
}

// authBindJSON json参数绑定至结构体
//
// paramJson 	参数。
// obj 			结构体。
// isAuthSign 	是否签名验证。
func (c *Context) authBindJSON(paramJson []byte, obj interface{}, authType bool) (ucUid int64, err error) {
	dataMap := make(map[string]interface{})
	if err = json.Unmarshal(paramJson, &dataMap); err != nil {
		logger.LogContext(c.Context, err).Warn("未知的参数格式")
		if config.AppMode() != "release" {
			return 0, fmt.Errorf("未知的参数格式，err：%v", err)
		}
		return 0, errors.New("未知的参数格式")
	}

	// 记录入参
	logger.LogContext(c.Context, dataMap).Info("入参")

	if err = json.Unmarshal(paramJson, obj); err != nil {
		logger.LogContext(c.Context, err).Warn("参数解析失败")
		if config.AppMode() != "release" {
			return 0, fmt.Errorf("参数解析失败，err：%v", err)
		}
		return 0, errors.New("参数解析失败")
	}

	if authType {
		if err = util.AuthSign(dataMap); err != nil {
			return 0, err
		}
	}

	if err = validate.Struct(obj); err != nil {
		logger.LogContext(c.Context, err).Warn("参数错误")
		if config.AppMode() != "release" {
			return 0, fmt.Errorf("参数错误，err：%v", err)
		}
		return 0, errors.New("参数错误")
	}

	if structField, ok := reflect.TypeOf(obj).Elem().FieldByName("Ut"); ok {
		JWTtoken, JWTerr := util.JWTParseToken(
			reflect.ValueOf(obj).Elem().FieldByName("Ut").String(),
			[]byte(config.AppSecret.AesAccountSecret["default"]),
		)

		switch JWTtoken["ucuid"].(type) {
		case float64:
			ucUid = int64(dataMap["ucuid"].(float64))
			break
		case string:
			ucUid, _ = strconv.ParseInt(dataMap["ucuid"].(string), 10, 64)
			break
		}
		if structField.Tag.Get("validate") == "required" && (ucUid == 0 || JWTerr != nil) {
			return 0, errors.New("请重新登录后再试")
		}
	}

	return ucUid, nil
}

// AuthBodyBindJSON body参数绑定至结构体，同一请求无法重复使用，
// 当结构体包含Ut时会进行Ut的认证，注意t是小写的。当Ut为必填时，会进行ucuid的认证。
// 参数认证方式参考：https://github.com/go-playground/validator
//
// obj 			结构体。
// authType 	签名验证类型，0不做验证，1Sign
func (c *Context) AuthBodyBindJSON(obj interface{}, isAuthSign bool) (ucUid int64, err error) {
	body, _ := ioutil.ReadAll(c.Request.Body)
	if len(body) == 0 {
		return 0, errors.New("未获取到body数据")
	}
	return c.authBindJSON(body, obj, isAuthSign)
}

// AuthAesEcbBodyBindJSON 	是否签名验证。
//
// obj 			结构体。
// isAuthSign 	是否签名验证。
// aesKey 		非必传，默认使用默认的rygSecret。
func (c *Context) AuthAesEcbBodyBindJSON(obj interface{}, isAuthSign bool, aesKey ...string) (ucUid int64, err error) {
	body, _ := ioutil.ReadAll(c.Request.Body)
	if len(body) == 0 {
		return 0, errors.New("未获取到body数据")
	}

	logger.LogContext(c.Context, string(body)).Info("加密入参")

	var key string
	if len(aesKey) > 0 && aesKey[0] != "" {
		key = config.AppSecret.AesAccountSecret[aesKey[0]]
	} else {
		key = config.AppSecret.AesAccountSecret["default"]
	}

	if c.Param("source") == "IOS" { // IOS数据特殊处理
		bodyStr, err := url.QueryUnescape(string(body))
		if err != nil {
			return 0, errors.New("IOS数据转换失败")
		}
		body = bytes.TrimLeft([]byte(bodyStr), "=")
	}

	body, err = util.AesDecryptECBBase64(body, []byte(key))
	if err != nil {
		return 0, errors.New("数据解密失败")
	}

	return c.authBindJSON(body, obj, isAuthSign)
}

/************************************** 上：入参先关 ******* 下：出参相关 ***********************************************/

// packMap 	组装参数
//
// errCode 	错误码，统一0正常，其他异常
// errMsg 	错误信息提示
// data 	数据内容，无内容传nil。
func (c *Context) packMap(errCode int, errMsg interface{}, data interface{}) map[string]interface{} {
	if data == nil {
		data = map[string]interface{}{}
	}
	switch errMsg.(type) {
	case error:
		errMsg = errMsg.(error).Error()
		break
	case []byte:
		errMsg = string(errMsg.([]byte))
		break
	}
	res := gin.H{
		"errcode": errCode,
		"errmsg":  errMsg,
		"runTime": util.GetRuntime(c.GetTime("StartTime")),
		"data":    data,
	}
	logger.LogContext(c.Context, res).Info("出参")
	if res["runTime"].(float64) >= config.App.SlowReqThreshold {
		c.ErrPush(errors.New(fmt.Sprintf("有一条慢请求记录，执行时长：%f秒", res["runTime"])))
	}
	return res
}

// Json 	返回json数据
//
// errCode 	错误码，统一0正常，其他异常
// errMsg 	错误信息提示
// data 	数据内容，无内容传nil。
func (c *Context) Json(errCode int, errMsg interface{}, data interface{}) *Context {
	c.JSON(http.StatusOK, c.packMap(errCode, errMsg, data))
	c.Abort()
	return c
}

// JsonAesEcb 	返回json后AesEcb加密数据
//
// errCode 	错误码，统一0正常，其他异常
// errMsg 	错误信息提示
// data 	数据内容，无内容传nil。
func (c *Context) JsonAesEcb(errCode int, errMsg interface{}, data interface{}, aesKey ...string) *Context {
	resJson, err := json.Marshal(c.packMap(errCode, errMsg, data))
	if err != nil {
		c.String(http.StatusInternalServerError, "返回数据编码错误")
		c.Abort()
		return c
	}
	var key string
	if len(aesKey) > 0 && aesKey[0] != "" {
		key = config.AppSecret.AesAccountSecret[aesKey[0]]
	} else {
		key = config.AppSecret.AesAccountSecret["default"]
	}
	resJsonAes, _ := util.AesEncryptECBBase64(resJson, []byte(key))
	c.String(http.StatusOK, string(resJsonAes))
	c.Abort()
	return c
}

// Log 记录上下文日志信息
func (c *Context) Log(logData interface{}) *zap.Logger {
	return logger.LogContext(c.Context, logData)
}

// ErrPush 记录并推送重要的更加详细的错误信息
// 可自定义推送至提醒平台，比如：钉钉、企业微信、邮箱
// 之所以在单独传递一次错误信息是因为可能返回给前端的数据与实际要记录的数据不一致
//
// errMsg 详细错误信息
func (c *Context) ErrPush(errMsg error) *Context {
	logger.ErrPush(c.Context, errMsg)
	return c
}
