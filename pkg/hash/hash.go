package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

type Hasher struct{ key string }

func NewHasher(key string) *Hasher {
	return &Hasher{key: key}
}

func (hsh *Hasher) GetHash(pwd string) string {
	h := hmac.New(sha256.New, []byte(hsh.key))
	h.Write([]byte(pwd))
	hash := fmt.Sprintf("%x", h.Sum(nil))
	return hash
}
