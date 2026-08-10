package campaign

import (
	"crypto/sha256"
	"encoding/hex"
)

func SHA256(body []byte) string {
	digest := sha256.Sum256(body)
	return "sha256:" + hex.EncodeToString(digest[:])
}
