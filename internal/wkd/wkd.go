package wkd

import (
    "crypto/sha1"
    "strings"
	"github.com/tv42/zbase32"
)

func WKDHash(email string) string {
    normalized := strings.ToLower(email)
    sum := sha1.Sum([]byte(normalized))
    return zbase32.EncodeToString(sum[:])
}
