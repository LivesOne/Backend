package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"utils/logger"
)

func Base64Encode(in []byte) string {
	return base64.StdEncoding.EncodeToString(in)
}

func Base64Decode(in string) []byte {
	decoded, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		logger.Info("decode base64 error:", err, in)
		return nil
	}

	return decoded
}

func HexEncode(in []byte) string {
	return hex.EncodeToString(in)
}

func HexDecode(in string) []byte {
	decoded, err := hex.DecodeString(in)
	if err != nil {
		logger.Info("decode hex to string failed", err, in)
		return nil
	}

	return decoded
}

func Sha256(in string) string {
	encode := sha256.Sum256([]byte(in))
	return HexEncode(encode[:])
}

//convert binary to base64 encoded string as output
func AesEncrypt(src, key, iv string) string {

	block, err := aes.NewCipher([]byte(key)) //选择加密算法
	if err != nil {
		logger.Info("AesEncrypt error", err)
		return ""
	}

	plaintText := pkcs7Padding(src, block.BlockSize())
	blockModel := cipher.NewCBCEncrypter(block, []byte(iv))
	ciphertext := make([]byte, len(plaintText))
	blockModel.CryptBlocks(ciphertext, plaintText)

	return Base64Encode(ciphertext)
}

//convert base64 encoded string to binary as input
func AesDecrypt(src /*base64 encoded string*/, key, iv string) string {

	keyBytes := Base64Decode(src)
	block, err := aes.NewCipher(keyBytes) //选择加密算法
	if err != nil {
		return ""
	}

	blockModel := cipher.NewCBCDecrypter(block, []byte(iv))
	plaintText := make([]byte, len(keyBytes))
	blockModel.CryptBlocks(plaintText, keyBytes)
	plaintText = pkcs7UnPadding(plaintText, block.BlockSize())

	return string(plaintText)
}

func pkcs7Padding(src string, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append([]byte(src), padtext...)
}

func pkcs7UnPadding(plantText []byte, blockSize int) []byte {
	length := len(plantText)
	unpadding := int(plantText[length-1])
	return plantText[:(length - unpadding)]
}

//convert base64 encoded string to binary as input
func RsaDecrypt(src /*base64 encoded string*/ string, privateKey string) string {
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		return ""
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return ""
	}

	decoded, err := rsa.DecryptPKCS1v15(rand.Reader, priv, Base64Decode(src))
	if err != nil {
		return ""
	}
	return string(decoded)
}

//convert binary to base64 encoded string as output
func RsaSign(src string, privateKey string) string {
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		return ""
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return ""
	}
	pub := pubInterface.(*rsa.PublicKey)
	encode, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(src))
	if err != nil {
		return ""
	}
	return Base64Encode(encode)
}
