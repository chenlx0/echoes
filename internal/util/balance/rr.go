package balance

import (
	"github.com/chenlx0/echoes/internal/config"
)

// RoundRobin implement LoadBlanceFunc,
func RoundRobin(ups []config.Upstream) *config.Upstream {
	return &ups[0]
}
