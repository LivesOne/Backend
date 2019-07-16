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

// AesEncrypt
// [in] src : original data
// [out] encryped data with base64 encode
func AesEncrypt(src, key, iv string) (string, error) {
	if (len(src) < 1) || (len(key) < 1) || (len(iv) < 1) {
		return "", errors.New("invalid param")
	}

	encrypt, err := aesEncrypt([]byte(src), []byte(key), []byte(iv))
	if err != nil {
		return "", err
	}
	return Base64Encode(encrypt), nil
}

func aesEncrypt(plaintext []byte, key []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if (err != nil) || (block == nil) {
		logger.Info("invalid decrypt key")
		return nil, errors.New("invalid decrypt key")
	}
	blockSize := block.BlockSize()
	plaintextNew := PKCS5Padding(plaintext, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(plaintextNew))
	blockMode.CryptBlocks(ciphertext, plaintextNew)

	return ciphertext, nil
}

// AesDecrypt
// [in] src : encryped data with base64 encode
// [out] original data
func AesDecrypt(src, key, iv string) (string, error) {
	if (len(src) < 1) || (len(key) < 1) || (len(iv) < 1) {
		return "", errors.New("invalid param")
	}

	srcDecode := Base64Decode(src)
	orig, err := aesDecrypt(srcDecode, []byte(key), []byte(iv))
	if err != nil {
		return "", err
	}
	return string(orig), nil
}

func aesDecrypt(src []byte, key []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if (err != nil) || (block == nil) {
		logger.Info("invalid decrypt key")
		return nil, errors.New("invalid decrypt key")
	}

	blockSize := block.BlockSize()
	if len(src) < blockSize {
		logger.Info("ciphertext too short")
		return nil, errors.New("ciphertext too short")
	}

	if len(src)%blockSize != 0 {
		logger.Info("ciphertext is not a multiple of the block size")
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	blockModel := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(src))
	blockModel.CryptBlocks(plaintext, src)
	plaintext = PKCS5UnPadding(plaintext)

	return plaintext, nil
}

func PKCS5Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	// if padding < 1 {
	// 	return src
	// }
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func PKCS5UnPadding(src []byte) []byte {
	length := len(src)
	if length < 1 {
		return src
	}
	unpadding := int(src[length-1])
	lenNew := length - unpadding
	if (lenNew < 0) || (lenNew > length-1) {
		return src
	}
	return src[:lenNew]
}

//convert base64 encoded string to binary as input
/*base64 encoded string*/
func RsaDecrypt(src string, privateKey []byte) (string, error) {

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




//AES ECB模式的加密解密
type AesTool struct {
	//128 192  256位的其中一个 长度 对应分别是 16 24  32字节长度
	Key       []byte
	BlockSize int
}

func NewAesTool(key []byte, blockSize int) *AesTool {
	return &AesTool{Key: key, BlockSize: blockSize}
}

func (this *AesTool) padding(src []byte) []byte {
	//填充个数
	paddingCount := aes.BlockSize - len(src)%aes.BlockSize
	if paddingCount == 0 {
		return src
	} else {
		//填充数据
		return append(src, bytes.Repeat([]byte{byte(0)}, paddingCount)...)
	}
}

//unpadding
func (this *AesTool) unPadding(src []byte) []byte {
	for i := len(src) - 1; ; i-- {
		if src[i] != 0 {
			return src[:i+1]
		}
	}
	return nil
}

func (this *AesTool) Encrypt(src []byte) ([]byte, error) {
	//key只能是 16 24 32长度
	block, err := aes.NewCipher([]byte(this.Key))
	if err != nil {
		logger.Error("aes ecb encrypt error",err.Error())
		return nil, err
	}
	//padding
	src = this.padding(src)
	//返回加密结果
	encryptData := make([]byte, len(src))
	//存储每次加密的数据
	tmpData := make([]byte, this.BlockSize)

	//分组分块加密
	for index := 0; index < len(src); index += this.BlockSize {
		block.Encrypt(tmpData, src[index:index+this.BlockSize])
		copy(encryptData, tmpData)
	}
	return encryptData, nil
}
func (this *AesTool) Decrypt(src []byte) ([]byte, error) {
	//key只能是 16 24 32长度
	block, err := aes.NewCipher([]byte(this.Key))
	if err != nil {
		return nil, err
	}
	//返回加密结果
	decryptData := make([]byte, len(src))
	//存储每次加密的数据
	tmpData := make([]byte, this.BlockSize)

	//分组分块加密
	for index := 0; index < len(src); index += this.BlockSize {
		block.Decrypt(tmpData, src[index:index+this.BlockSize])
		copy(decryptData, tmpData)
	}
	return this.unPadding(decryptData), nil
}