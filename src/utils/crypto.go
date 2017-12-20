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
	"errors"
	"utils/logger"
	"utils/lvtrsa"
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
func AesEncrypt(src, key, iv string) (string, error) {

	block, err := aes.NewCipher([]byte(key)) //选择加密算法
	if err != nil {
		logger.Info("AesEncrypt error", err)
		return "", err
	}

	plaintText := pkcs7Padding(src, block.BlockSize())
	blockModel := cipher.NewCBCEncrypter(block, []byte(iv))
	ciphertext := make([]byte, len(plaintText))
	blockModel.CryptBlocks(ciphertext, plaintText)

	return Base64Encode(ciphertext), nil
}

//convert base64 encoded string to binary as input
// /*base64 encoded string*/
func AesDecrypt(src, key, iv string) (string, error) {

	keyBytes := Base64Decode(src)
	block, err := aes.NewCipher(keyBytes) //选择加密算法
	if err != nil {
		return "", err
	}

	blockModel := cipher.NewCBCDecrypter(block, []byte(iv))
	plaintText := make([]byte, len(keyBytes))
	blockModel.CryptBlocks(plaintText, keyBytes)
	plaintText = pkcs7UnPadding(plaintText, block.BlockSize())

	return string(plaintText), nil
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
/*base64 encoded string*/
// func RsaDecrypt(src string, privateKey string) (string, error) {
// block, _ := pem.Decode([]byte(privateKey))
func RsaDecrypt(src string, privateKey []byte) (string, error) {

	// RsaSign("slic客户端和Xebo服务器之间通信过程中加密敏感 信息.", privateKey)

	block, _ := pem.Decode(privateKey)
	if block == nil {
		// fmt.Println("AAAAAAAAAAAAAAAAAAAAAAAAAA no pem data")
		return "", errors.New("no pem data")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	// base64Origin := Base64Decode(src)
	// fmt.Println("bbbbbbbbbbbbbbbb, base64Origin:", src, string(base64Origin))
	// decoded, err := rsa.DecryptPKCS1v15(rand.Reader, priv, base64Origin)
	decoded, err := rsa.DecryptPKCS1v15(rand.Reader, priv, Base64Decode(src))
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// RsaSign sign the src with private key
func RsaSign(src string, privateKeyFilename string) (string, error) {
	brsa, err := lvtrsa.PrivateEncrypt([]byte(src), privateKeyFilename, lvtrsa.RSA_PKCS1_PADDING)
	if err != nil {
		logger.Info("RsaSign error %s\n", err)
		return "", err
	}

	base64encoded := Base64Encode(brsa)
	logger.Info("-----------base64:", base64encoded)
	return base64encoded, nil
}

// func RsaSignOld(src string, privateKey []byte) (string, error) {
// 	block, _ := pem.Decode(privateKey)
// 	if block == nil {
// 		fmt.Println("AAAAAAAAAAAAAAAAAAAAAAAAAA no pem data")
// 		return "", errors.New("no pem data")
// 	}
// 	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
// 	// priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
// 	if err != nil {
// 		fmt.Println("bbbbbbbbbbbbbbbbb ")
// 		return "", err
// 	}
// 	pub := pubInterface.(*rsa.PublicKey)
// 	encode, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(src))
// 	if err != nil {
// 		return "", err
// 	}
// 	logger.Info("=============, aaaaaaaaaaaaaaa", Base64Encode(encode))
// 	return Base64Encode(encode), nil
// }
