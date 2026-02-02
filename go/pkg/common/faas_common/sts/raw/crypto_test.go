package raw

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	saltKeySep = ":"

	shareKey = "1752F862B5176946F18D45D67E256642F115D2D6A3D77773FAF1E5874AC5211D"

	plain = "{\"key1\":\"value1\",\"key2\":\"value2\"}"

	saltKey = "R8Mi3gSG3ou4X6eY:VIQASOEBJTQT3yd4qGrpqSbLrgemB5eTaD5KRefaOcXh/r18YSwhtv0j0A=="
	plain2  = "{\"key1\":\"va1\",\"key2\":\"va2\"}"
)

func TestAesGCMDecrypt(t *testing.T) {
	shareKey2 := make([]byte, hex.DecodedLen(len(shareKey)))
	_, err := hex.Decode(shareKey2, []byte(shareKey))
	if err != nil {
		t.Errorf("%s", err)
	}

	salt, cipherBytes, err := AesGCMEncrypt(shareKey2, []byte(plain))
	fmt.Println(string(salt), string(cipherBytes))
	saltBase64 := base64.StdEncoding.EncodeToString(salt)
	cipherBase64 := base64.StdEncoding.EncodeToString(cipherBytes)
	fmt.Println(saltBase64, cipherBase64)

	if err != nil {
		t.Errorf("%s", err)
	}
	blocks1, err := AesGCMDecrypt(shareKey2, salt, cipherBytes)

	assert.Equal(t, string(blocks1), plain)
}

func TestAesGCMDecrypt2(t *testing.T) {
	shareKey2 := make([]byte, hex.DecodedLen(len(shareKey)))
	_, err := hex.Decode(shareKey2, []byte(shareKey))
	if err != nil {
		t.Errorf("%s", err)
	}

	fields := strings.Split(saltKey, saltKeySep)
	salt1, err := base64.StdEncoding.DecodeString(fields[0])
	if err != nil {
		t.Errorf("%s", err)
	}
	cipher, err := base64.StdEncoding.DecodeString(fields[1])
	blocks, err := AesGCMDecrypt(shareKey2, salt1, cipher)
	if err != nil {
		t.Errorf("%s", err)
	}

	assert.Equal(t, string(blocks), plain2)
}
