package util

import (
	"bytes"
	cryptoRand "crypto/rand"
	"math/big"
	"math/rand"
	"os"
	"time"
)

func Tick(Milliseconds int64, f func()) chan bool {
	ticker := time.NewTicker(time.Millisecond * time.Duration(Milliseconds))
	stopChan := make(chan bool)
	go func(ticker *time.Ticker) {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				f()
			case stop := <-stopChan:
				if stop {
					return
				}
			}
		}
	}(ticker)
	return stopChan
}

func TickStop(c chan bool) {
	c <- true
	close(c)
}

// Exists 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// GetRuntime 获取间隔时间
func GetRuntime(t time.Time) float64 {
	return time.Now().Sub(t).Seconds()
}

// RandomString 生成随机数
// size 随机码的位数
// kind 0纯数字、1小写字母、2大写字母、3数字+大小写字母
func RandomString(size int, kind int) string {
	ikind, kinds, rsbytes := kind, [][]int{[]int{10, 48}, []int{26, 97}, []int{26, 65}}, make([]byte, size)
	isAll := kind > 2 || kind < 0
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < size; i++ {
		if isAll { // random ikind
			ikind = rand.Intn(3)
		}
		scope, base := kinds[ikind][0], kinds[ikind][1]
		rsbytes[i] = uint8(base + rand.Intn(scope))
	}
	return string(rsbytes)
}

// CreateRandomString 生成随机数
func CreateRandomString(len int) string {
	var container string
	var str = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := bytes.NewBufferString(str)
	bigInt := big.NewInt(int64(b.Len()))
	for i := 0; i < len; i++ {
		randomInt, _ := cryptoRand.Int(cryptoRand.Reader, bigInt)
		container += string(str[randomInt.Int64()])
	}
	return container
}
