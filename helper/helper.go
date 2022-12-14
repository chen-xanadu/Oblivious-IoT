package helper

import (
	"Oblivious-IoT/config"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
)

func ReadPk(filename string) *rsa.PublicKey {
	_, curPath, _, _ := runtime.Caller(0)
	filename = path.Join(path.Dir(curPath), "/../", filename)

	pkRaw, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("error when reading pk file: %s \n", filename)
		panic(err)
	}

	block, _ := pem.Decode(pkRaw)
	if block == nil || block.Type != "RSA_PUBLIC_KEY" {
		fmt.Printf("error when decoding pk: %s \n", filename)
		panic(err)
	}

	pk, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		fmt.Printf("error when parsing pk: %s \n", filename)
		panic(err)
	}

	return pk
}

func ReadSk(filename string) *rsa.PrivateKey {
	_, curPath, _, _ := runtime.Caller(0)
	filename = path.Join(path.Dir(curPath), "/../", filename)

	skRaw, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("error when reading sk file: %s \n", filename)
		panic(err)
	}

	block, _ := pem.Decode(skRaw)
	if block == nil || block.Type != "RSA_PRIVATE_KEY" {
		fmt.Printf("error when decoding sk: %s \n", filename)
		panic(err)
	}

	sk, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		fmt.Printf("error when parsing sk: %s \n", filename)
		panic(err)
	}

	return sk
}

func GenerateEid(j, rid int) []byte {
	prf := hmac.New(sha256.New, config.DeviceHmacKey)

	data := binary.LittleEndian.AppendUint32(nil, uint32(j))
	data = binary.LittleEndian.AppendUint32(data, uint32(rid))

	prf.Write(data)

	return prf.Sum(nil)
}

func OnionEncrypt(m []byte, pks []*rsa.PublicKey) []byte {
	data := m
	for _, pk := range pks {
		data = HybridEncrypt(data, pk)
	}
	return data
}

func HybridEncrypt(plaintext []byte, pk *rsa.PublicKey) []byte {
	aesKey := make([]byte, 32)
	_, _ = rand.Read(aesKey)

	ciphertext := CBCEncrypt(plaintext, aesKey)

	aesKeyEnc, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pk, aesKey, nil)
	if err != nil {
		panic(err)
	}

	return append(aesKeyEnc, ciphertext...)
}

func HybridDecrypt(ciphertext []byte, sk *rsa.PrivateKey) []byte {
	aesKeyEnc := ciphertext[:256]
	ciphertext = ciphertext[256:]

	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, sk, aesKeyEnc, nil)
	if err != nil {
		panic(err)
	}

	plaintext := CBCDecrypt(ciphertext, aesKey)

	return plaintext
}

func CBCEncrypt(plaintext []byte, key []byte) []byte {
	bPlaintext := PKCS5Padding(plaintext, aes.BlockSize, len(plaintext))
	block, _ := aes.NewCipher(key)
	ciphertext := make([]byte, aes.BlockSize+len(bPlaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], bPlaintext)
	return ciphertext
}

func CBCDecrypt(ciphertext []byte, key []byte) []byte {
	block, _ := aes.NewCipher(key)

	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	if len(ciphertext)%aes.BlockSize != 0 {
		panic("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	mode.CryptBlocks(ciphertext, ciphertext)
	return ciphertext
}

func PKCS5Padding(ciphertext []byte, blockSize int, after int) []byte {
	padding := (blockSize - len(ciphertext)%blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//func NewPermutation(n int) []int {
//	indices := make([]int, n)
//	for i := range indices {
//		indices[i] = i
//	}
//	rand2.Shuffle(n, func(i, j int) {
//		fmt.Println(i, j)
//		indices[i], indices[j] = indices[j], indices[i]
//	})
//	return indices
//}
