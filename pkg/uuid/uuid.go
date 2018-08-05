package uuid

import (
	"crypto/rand"
	"encoding/hex"
	"io"
)

func NewUUID() string {
	uuid := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, uuid)
	if err != nil {
		// This can't happen unless the system is badly
		// configured, like /dev/urandom isn't readable.
		panic("reading random: " + err.Error())
	}

	return stringify(uuid)
}

// stringify returns canonical string representation of UUID:
// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
func stringify(uuid []byte) string {
	buf := make([]byte, 36)

	hex.Encode(buf[0:8], uuid[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], uuid[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], uuid[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], uuid[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], uuid[10:])

	return string(buf)
}
