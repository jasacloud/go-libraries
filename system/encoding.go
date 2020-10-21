// Copyright (c) 2019 JasaCloud.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package system

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"strings"
)

const (
	asccii64          = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	asccii64url       = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	asccii64custom    = "/ABCDEFHIJKLMNOPQRUVWXYZGSTabcdefhijkl+mnopqruvwxyzgst0234678915"
	asccii64urlcustom = "_ABCDEFHIJKLMNOPQRUVWXYZGSTabcdefhijkl-mnopqruvwxyzgst0234678915"
)

// Base64encode function
func Base64encode(msg string) string {

	return base64.StdEncoding.EncodeToString([]byte(msg))
}

// Base64decode function
func Base64decode(msg string) string {
	decoded, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		return ""
	}

	return string(decoded)
}

// Base64urlencode function
func Base64urlencode(msg string) string {

	return base64.URLEncoding.EncodeToString([]byte(msg))
}

// Base64urldecode function
func Base64urldecode(msg string) string {
	decoded, err := base64.URLEncoding.DecodeString(msg)
	if err != nil {
		return ""
	}

	return string(decoded)
}

// Base64urlencodeCustom function
func Base64urlencodeCustom(msg string) string {
	arr := []string{}
	encode := base64.URLEncoding.EncodeToString([]byte(msg))
	for i, v := range asccii64url {
		arr = append(arr, string(v), string(asccii64urlcustom[i]))
	}
	r := strings.NewReplacer(arr...)
	result := r.Replace(encode)

	return result
}

// Base64urldecodeCustom function
func Base64urldecodeCustom(msg string) string {
	arr := []string{}
	for i, v := range asccii64url {
		arr = append(arr, string(asccii64urlcustom[i]), string(v))
	}
	r := strings.NewReplacer(arr...)
	msg = r.Replace(msg)

	decoded, err := base64.URLEncoding.DecodeString(msg)
	if err != nil {
		return ""
	}

	return string(decoded)
}

// Base64encodeCustom function
func Base64encodeCustom(msg string) string {
	arr := []string{}
	encode := base64.StdEncoding.EncodeToString([]byte(msg))
	for i, v := range asccii64 {
		arr = append(arr, string(v), string(asccii64custom[i]))
	}
	r := strings.NewReplacer(arr...)
	result := r.Replace(encode)

	return result
}

// Base64decodeCustom function
func Base64decodeCustom(msg string) string {
	arr := []string{}
	for i, v := range asccii64 {
		arr = append(arr, string(asccii64custom[i]), string(v))
	}
	r := strings.NewReplacer(arr...)
	msg = r.Replace(msg)

	decoded, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		return ""
	}

	return string(decoded)
}

// Md5 function
func Md5(msg string) string {
	hasher := md5.New()
	hasher.Write([]byte(msg))

	return hex.EncodeToString(hasher.Sum(nil))
}

// TripleDesEncrypt function (3DES)
func TripleDesEncrypt(data, key []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	// note: IV = key[:8]
	blockMode := cipher.NewCBCDecrypter(block, key[:8])
	origData := make([]byte, len(data))
	// origData := crypted
	blockMode.CryptBlocks(origData, data)
	return origData, nil
}

// TripleDesEncryptString function (3DES)
func TripleDesEncryptString(data, keyString string) ([]byte, error) {
	crypted := []byte(data)
	key := []byte(keyString)
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	// note: IV = key[:8]
	blockMode := cipher.NewCBCDecrypter(block, key[:8])
	origData := make([]byte, len(crypted))
	// origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	return origData, nil
}

// TripleDesDecrypt function (3DES)
func TripleDesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	// note: IV = key[:8]
	blockMode := cipher.NewCBCDecrypter(block, key[:8])
	origData := make([]byte, len(crypted))
	// origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	return origData, nil
}

// TripleDesDecryptString function (3DES)
func TripleDesDecryptString(cryptedString, keyString string) ([]byte, error) {
	crypted := []byte(cryptedString)
	key := []byte(keyString)
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	// note: IV = key[:8]
	blockMode := cipher.NewCBCDecrypter(block, key[:8])
	origData := make([]byte, len(crypted))
	// origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	return origData, nil
}

// Sha256String function
func Sha256String(s string) string {
	h := sha256.New()
	h.Write([]byte(s))

	return hex.EncodeToString(h.Sum(nil))
}

// HMACSha256 function
func HMACSha256(secret, stringToSign string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(stringToSign))

	return hex.EncodeToString(mac.Sum(nil))
}

// EncryptAes function
func EncryptAes(key []byte, message string) (encmess string, err error) {
	plainText := []byte(message)

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	//returns to base64 encoded string
	encmess = base64.StdEncoding.EncodeToString(cipherText)
	return
}

// DecryptAes function
func DecryptAes(key []byte, securemess string) (decodedmess string, err error) {
	cipherText, err := base64.StdEncoding.DecodeString(securemess)
	if err != nil {
		return
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	if len(cipherText) < aes.BlockSize {
		err = errors.New("Ciphertext block size is too short!")
		return
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(cipherText, cipherText)

	decodedmess = string(cipherText)
	return
}
