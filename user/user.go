package user

import (
	"Oblivious-IoT/config"
	"Oblivious-IoT/helper"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"strings"
	"time"
)

type Command struct {
	Name      string
	Timestamp int64
}

func (cmd Command) encrypt() []byte {
	devPk := helper.ReadPk(config.DevicePkFile)

	data := cmd.serialize()

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, devPk, data, nil)
	if err != nil {
		fmt.Printf("error when encrypting %v \n", cmd)
		panic(err)
	}

	return ciphertext
}

func (cmd Command) serialize() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(cmd)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

type UserMessage struct {
	Cmd []byte
	Eid []byte
	Vid []byte
}

func (m *UserMessage) Serialize() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(&m)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func (m *UserMessage) Deserialize(data []byte) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	err := dec.Decode(m)
	if err != nil {
		panic(err)
	}

	return
}

func GenerateUserRequest(input string) []byte {
	plainCmd := Command{strings.Repeat(input, 10), time.Now().Unix()}

	integratorPk := helper.ReadPk(config.IntegratorPkFile)
	vendorPk := helper.ReadPk(config.VendorPkFile)

	m := &UserMessage{
		Cmd: plainCmd.encrypt(),
		Eid: helper.GenerateEid(1, config.RoundID),
		Vid: binary.LittleEndian.AppendUint32(nil, config.VendorID),
	}

	pk_ring := []*rsa.PublicKey{integratorPk, vendorPk}

	return helper.OnionEncrypt(m.Serialize(), pk_ring)
}
