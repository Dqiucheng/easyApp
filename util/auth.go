package util

import (
	"crypto/md5"
	"easyApp/config"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

// JWTGenerateToken 生成token
func JWTGenerateToken(claims jwt.MapClaims, secret []byte) (tokenString string, err error) {
	// 创建一个新的令牌对象，指定签名方法和声明
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用密码签名并获得完整的编码令牌作为字符串
	tokenString, err = token.SignedString(secret)
	return
}

// JWTParseToken 解析token
func JWTParseToken(tokenString string, secret []byte) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err == nil {
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			return claims, nil
		}
	}

	return nil, err
}

// GetSign 生成Sign
func GetSign(m map[string]interface{}) string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys) // 升序排序
	builder := strings.Builder{}
	MD5builder := strings.Builder{}
	for _, v := range keys {
		builder.WriteString(v)
		builder.WriteString("=")
		switch m[v].(type) {
		case float64:
			builder.WriteString(strconv.FormatFloat(m[v].(float64), 'f', -1, 64))
			break
		default:
			builder.WriteString(fmt.Sprint(m[v]))
			break
		}
		builder.WriteString("&")
	}

	sign := builder.String()
	MD5builder.WriteString(fmt.Sprintf("%x", md5.Sum([]byte(sign[:len(sign)-1])))) // 排序后去除尾部特殊字符进行MD5后在拼接
	MD5builder.WriteString(fmt.Sprint(m["nonce_str"]))
	MD5builder.WriteString(fmt.Sprint(m["timestamp"]))
	MD5builder.WriteString(
		config.AppSecret.SignAccountSecret[fmt.Sprint(m["appkey"])], // 拼接appSecret
	)
	return strings.ToUpper(fmt.Sprintf("%x", md5.Sum([]byte(MD5builder.String()))))
}

// AuthSign 验证Sign
func AuthSign(paramMap map[string]interface{}) error {
	switch paramMap["timestamp"].(type) {
	case float64:
		paramMap["timestamp"] = int64(paramMap["timestamp"].(float64))
		break
	case string:
		paramMap["timestamp"], _ = strconv.ParseInt(paramMap["timestamp"].(string), 10, 64)
		break
	case nil:
		return errors.New("[timestamp]错误")
	}
	if math.Abs(float64(paramMap["timestamp"].(int64)-time.Now().Unix())) > 120 {
		return errors.New("签名过期，请校对设备时间与标准时间保持一致后再试")
	}

	sign, ok := paramMap["sign"].(string)
	if ok {
		delete(paramMap, "sign")
		if sign != GetSign(paramMap) {
			return errors.New("签名错误")
		}
	} else {
		return errors.New("[sign]无法识别")
	}
	return nil
}
