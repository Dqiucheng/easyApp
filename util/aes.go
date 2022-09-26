package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

// AesEncryptCBC =================== CBC ======================
func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
func AesEncryptCBC(origData []byte, key []byte) (encrypted []byte) {
	// 分组秘钥
	// NewCipher该函数限制了输入k的长度必须为16, 24或者32
	block, _ := aes.NewCipher(key)
	blockSize := block.BlockSize()                              // 获取秘钥块的长度
	origData = pkcs5Padding(origData, blockSize)                // 补全码
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize]) // 加密模式
	encrypted = make([]byte, len(origData))                     // 创建数组
	blockMode.CryptBlocks(encrypted, origData)                  // 加密
	return encrypted
}
func AesDecryptCBC(encrypted []byte, key []byte) (decrypted []byte) {
	block, _ := aes.NewCipher(key)                              // 分组秘钥
	blockSize := block.BlockSize()                              // 获取秘钥块的长度
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize]) // 加密模式
	decrypted = make([]byte, len(encrypted))                    // 创建数组
	blockMode.CryptBlocks(decrypted, encrypted)                 // 解密
	decrypted = pkcs5UnPadding(decrypted)                       // 去除补全码
	return decrypted
}

// AesEncryptECB =================== ECB ======================
func AesEncryptECB(origData []byte, key []byte) ([]byte, error) {
	newCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	length := (len(origData) + aes.BlockSize) / aes.BlockSize
	plain := make([]byte, length*aes.BlockSize)
	copy(plain, origData)
	pad := byte(len(plain) - len(origData))
	for i := len(origData); i < len(plain); i++ {
		plain[i] = pad
	}
	encrypted := make([]byte, len(plain))
	// 分组分块加密
	for bs, be := 0, newCipher.BlockSize(); bs <= len(origData); bs, be = bs+newCipher.BlockSize(), be+newCipher.BlockSize() {
		newCipher.Encrypt(encrypted[bs:be], plain[bs:be])
	}

	return encrypted, nil
}
func AesDecryptECB(encrypted []byte, key []byte) ([]byte, error) {
	newCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	size := newCipher.BlockSize()
	decrypted := make([]byte, len(encrypted))

	for bs, be := 0, size; bs < len(encrypted); bs, be = bs+size, be+size {
		newCipher.Decrypt(decrypted[bs:be], encrypted[bs:be])
	}

	return decrypted[0 : len(decrypted)-int(decrypted[len(decrypted)-1])], nil // 删除填充并返回
}
func AesEncryptECBBase64(encrypted []byte, key []byte) ([]byte, error) {
	encrypted, err := AesEncryptECB(encrypted, key)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(encrypted)))
	base64.StdEncoding.Encode(buf, encrypted)
	return buf, nil
}
func AesDecryptECBBase64(encrypted []byte, key []byte) ([]byte, error) {
	dBuf := make([]byte, base64.StdEncoding.DecodedLen(len(encrypted)))
	n, err := base64.StdEncoding.Decode(dBuf, encrypted)
	if err != nil {
		return nil, err
	}
	return AesDecryptECB(dBuf[:n], key)
}

// AesEncryptCFB =================== CFB ======================
func AesEncryptCFB(origData []byte, key []byte) (encrypted []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	encrypted = make([]byte, aes.BlockSize+len(origData))
	iv := encrypted[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(encrypted[aes.BlockSize:], origData)
	return encrypted
}
func AesDecryptCFB(encrypted []byte, key []byte) (decrypted []byte) {
	block, _ := aes.NewCipher(key)
	if len(encrypted) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := encrypted[:aes.BlockSize]
	encrypted = encrypted[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(encrypted, encrypted)
	return encrypted
}
