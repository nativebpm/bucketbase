package pocketbase

import (
	"regexp"
)

func ValidateEncryptionKey(key string) bool {
	if len(key) != 32 {
		return false
	}
	matched, err := regexp.MatchString(`^[0-9a-fA-F]{32}$`, key)
	return err == nil && matched
}

