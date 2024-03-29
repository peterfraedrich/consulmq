package kvmq

import (
	"strings"

	"github.com/google/uuid"
)

func hash(item []byte) string {
	_ = item
	uid, _ := uuid.NewRandom()
	return strings.ReplaceAll(uid.String(), "-", "")
}
