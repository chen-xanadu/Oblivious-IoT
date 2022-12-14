package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

func initKeyPair(pkFilename, skFilename string) {
	// generate key pair
	sk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	pk := &sk.PublicKey

	// encode sk and save to file
	skFile, err := os.Create(skFilename)
	if err != nil {
		panic(err)
	}
	err = pem.Encode(skFile, &pem.Block{
		Type:  "RSA_PRIVATE_KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(sk),
	})
	if err != nil {
		panic(err)
	}

	// encode pk and save to file
	pkFile, err := os.Create(pkFilename)
	if err != nil {
		panic(err)
	}
	err = pem.Encode(pkFile, &pem.Block{
		Type:  "RSA_PUBLIC_KEY",
		Bytes: x509.MarshalPKCS1PublicKey(pk),
	})
	if err != nil {
		panic(err)
	}
}

func main() {
	//initKeyPair(config.DevicePkFile, config.DeviceSkFile)
	//initKeyPair(config.VendorPkFile, config.VendorSkFile)
	//initKeyPair(config.IntegratorPkFile, config.IntegratorSkFile)

	start := time.Now()

	for i := 0; i < 1000; i++ {
		prf := hmac.New(sha256.New, []byte("device1"))
		for j := 0; j < 10; j++ {
			data := binary.LittleEndian.AppendUint32(nil, uint32(j))
			data = binary.LittleEndian.AppendUint32(data, uint32(1))

			prf.Write(data)
			prf.Sum(nil)
			prf.Reset()
		}
	}

	duration := time.Since(start)
	fmt.Println(duration)

}
