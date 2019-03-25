package untils

import (
	"crypto/cipher"
	"crypto/aes"
	"bytes"
	"encoding/base64"
)

//解密
func AesDecryptSimple(origData string, key []byte, iv string) (res []byte,errs error) {
	if bs,err := base64.StdEncoding.DecodeString(origData);err==nil{
		if r,err:=AesDecryptPkcs5([]byte(bs), key, []byte(iv));err==nil{
			return r,nil
		}else{
			return res,err
		}
	}else{
		return res,err
	}
}

func AesEncryptPkcs5(origData []byte, key []byte, iv []byte ) ([]byte, error) {
	return AesEncrypt(origData, key, iv, PKCS5Padding)
}

func AesEncrypt(origData []byte, key []byte, iv []byte, paddingFunc func([]byte, int) []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = paddingFunc(origData, blockSize)

	blockMode := cipher.NewCBCEncrypter(block, iv)
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}
//解加密
func AesEncryptSimple(crypted []byte, key []byte, iv string) (res string,errs error) {
	if aseCode,err := AesEncryptPkcs5(crypted, key, []byte(iv));err==nil{
		bas := base64.StdEncoding.EncodeToString(aseCode)
		return bas,nil
	}else{
		return res,err
	}
}

func AesDecryptPkcs5(crypted []byte, key []byte, iv []byte) ([]byte, error) {
	return AesDecrypt(crypted, key, iv, PKCS5UnPadding)
}

func AesDecrypt(crypted, key []byte, iv []byte, unPaddingFunc func([]byte) []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = unPaddingFunc(origData)
	return origData, nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	if length < unpadding {
		return []byte("unpadding error")
	}
	return origData[:(length - unpadding)]
}