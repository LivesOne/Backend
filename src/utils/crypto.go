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
type ecb struct {
	b         cipher.Block
	blockSize int
}

func newECB(key []byte) *ecb {
	b,e := aes.NewCipher(key)
	if e != nil {
		logger.Error(e.Error())
		return nil
	}
	return &ecb{
		b:         b,
		blockSize: b.BlockSize(),
	}
}

type ecbEncrypter ecb

// NewECBEncrypter returns a BlockMode which encrypts in electronic code book
// mode, using the given Block.
func New128ECBEncrypter(key string) *ecbEncrypter {
	return (*ecbEncrypter)(newECB(paddingKey(key,128)))
}

func New256ECBEncrypter(key string) *ecbEncrypter {
	return (*ecbEncrypter)(newECB(paddingKey(key)))
}

func (x *ecbEncrypter) BlockSize() int { return x.blockSize }

func (x *ecbEncrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}
	for len(src) > 0 {
		x.b.Encrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

func  (x *ecbEncrypter) Crypt(src string)string{
	sc := PKCS5Padding([]byte(src),x.BlockSize())
	tg := make([]byte,len(sc))
	x.CryptBlocks(tg,sc)
	return Base64Encode(tg)
}

func paddingKey(key string,mode ...int)[]byte{
	const (
		t_128 = 16
		t_256 = 32
	)
	t := t_256
	if len(mode) > 0 {
		switch mode[0] {
		case 128:
			t = t_128
		case 256:
			t = t_256
		}
	}

	sc := []byte(key)
	if len(sc) >= t {
		return sc[:t]
	}
	padtext := bytes.Repeat([]byte{0x00}, t-len(sc))
	return append(sc,padtext...)
}