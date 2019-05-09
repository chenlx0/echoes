package balance

import (
	"crypto/md5"
	"encoding/binary"

	"github.com/chenlx0/echoes/internal/config"
)

type HashFunc func(string) *config.Upstream

// Hash implement LoadBlanceFunc,
func Hash(ups []config.Upstream) *config.Upstream {
	return &ups[0]
}

// GetHashFunc return hash func with upstreams
func GetHashFunc(upstreams []*config.Upstream) HashFunc {
	size := len(upstreams)
	return func(key string) *config.Upstream {
		sum := md5.Sum([]byte(key))
		hashNum := binary.BigEndian.Uint32(sum[:])
		index := int(hashNum) % size
		return upstreams[index]
	}
}
